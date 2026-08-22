package rendering

import (
	"container/heap"
	"errors"
	"fmt"

	"kaijuengine.com/matrix"
)

const (
	// meshQemChunkTargetVerts is the maximum number of vertices a single chunk may
	// hold before the mesh is split again. It bounds per-goroutine work so chunk
	// processing stays short enough to run at load/runtime without a perceptible
	// hitch.
	meshQemChunkTargetVerts = 4096
)

type MeshLODInstance struct {
	Mesh  *Mesh
	Ratio float32
}

type MeshLod struct {
	Levels []MeshLODInstance
}

// I'm too stupid to know this stuff off of the top of my head, so I'm going
// to place a block comment down below to help guide me with this implementation.

// Quadric Error Metric (QEM) is a technique used for mesh simplification (LOD
// generation). Each vertex is assigned a "quadric" — a small symmetric 4x4
// matrix that encodes the set of planes of the triangles surrounding that
// vertex. Given a vertex position v, the quadric Q yields the quadratic form
// v^T Q v, which approximates the sum of squared distances from v to each of
// those triangle planes. It therefore measures how much error would be
// introduced if the vertex were moved.
//
// Simplification proceeds by repeatedly collapsing the edge (v1, v2) whose
// combined quadric Q = Q1 + Q2 predicts the smallest error. The collapsed
// vertex is placed at the point that minimizes v^T Q v (the least-squares
// optimal position), and its quadric is replaced by the sum of the two
// collapsed vertices' quadrics so the error estimate is carried forward.
// This greedy, error-driven approach lets the algorithm remove geometry where
// it matters least visually, keeping the surface as faithful as possible for
// the requested triangle budget.

type MeshQem struct {
	chunks []MeshQemChunk
}

// BRENT NOTE - I am wanting to do this automatic LOD generation at runtime.
// It will allow for better game modding support from players. This does
// mean that I need to do this chunking thing, which has the potential to
// produce some issues across edge boundaries.

// MeshQemChunk is a self-contained, spatially-local slice of a mesh that can be
// simplified independently (typically on its own thread) after chunking.
//
// Chunking happens at runtime so that content loaded from player mods can be
// simplified on the fly without pre-baking LODs into the asset. Each chunk owns
// private copies of its geometry so that multiple chunks can be processed in
// parallel without sharing (and racing on) the source buffers.
//
// Inputs populated by chunkify:
//   - positions:       the chunk's vertices, compacted to [0, len(positions)).
//   - indices:         the chunk's triangles, remapped into local position space.
//   - boundary:        for each local position, whether it lies on the chunk's
//     outer border. Border vertices are shared with neighboring chunks and MUST
//     be held fixed during simplification so adjacent chunks stay watertight
//     and no seams appear when the results are stitched back together.
//   - globalIndices:   for each local position, the original vertex index it
//     came from in the full mesh. Used to weld chunks back together.
//   - targetTriangles: the triangle budget this chunk should collapse down to,
//     derived from the requested simplification ratio.
//
// Outputs populated by processChunk:
//   - quadrics:      the per-vertex QEM matrices (len == len(positions)).
//   - resultIndices: the simplified triangle list, referencing local position
//     indices, ready to be merged with the other chunks' results.
type MeshQemChunk struct {
	// Inputs
	positions       []matrix.Vec3
	indices         []uint32
	boundary        []bool
	globalIndices   []int
	targetTriangles int

	// Outputs
	quadrics      []matrix.Mat4
	resultIndices []uint32
}

// meshQemSlot records where a global vertex lives: in which chunk and at which
// local index within that chunk.
type meshQemSlot struct {
	chunk int
	local int
}

// qemEdgeHeap is a min-heap of qemEdge ordered by cost.
type qemEdgeHeap []*qemEdge

func (l MeshLod) IsValid() bool { return len(l.Levels) > 0 }

// GenerateMeshLOD builds a set of level-of-detail meshes for the given source
// mesh using Quadric Error Metric (QEM) simplification.
//
// The result is a MeshLod whose Level 0 is always the full-resolution source
// mesh, followed by levels 1..levels-1 whose triangle counts fall off
// power-of-two (see selectMeshLodRatios): level i has ratio 1.0/2^i.
//
// Generation happens at runtime (rather than being pre-baked offline) so that
// content loaded from player mods can be simplified on demand. The algorithm
// chunks the mesh (quadricErrorMetricChunkify), simplifies each chunk
// independently (quadricErrorMetricProcessChunk — parallelizable because
// boundary vertices are held fixed), and welds the chunk results back into a
// single watertight mesh (stitchQemChunks).
//
// The computed LODs are cached on mesh so repeated calls return the same set.
// verts and indices describe the full-resolution geometry; if a level cannot
// be reduced, that level falls back to reusing the source mesh itself.
func GenerateMeshLOD(mesh *Mesh, verts []Vertex, indices []uint32, levels int) (MeshLod, error) {
	if levels <= 0 {
		return MeshLod{}, errors.New("GenerateMeshLOD: levels must be greater than 0")
	}
	if mesh.lods.IsValid() {
		// Mesh has already been processed
		return mesh.lods, nil
	}
	lods := MeshLod{
		Levels: make([]MeshLODInstance, levels),
	}
	// Level 0 is always the full-resolution mesh itself.
	lods.Levels[0] = MeshLODInstance{Mesh: mesh, Ratio: 1}
	ratios := selectMeshLodRatios(levels)
	for i := 1; i < levels; i++ {
		ratio := ratios[i]
		chunks := quadricErrorMetricChunkify(verts, indices, ratio)
		if len(chunks) == 0 {
			// Nothing to simplify; reuse the full mesh for this level.
			lods.Levels[i] = MeshLODInstance{Mesh: mesh, Ratio: ratio}
			continue
		}
		// Simplify each chunk. This can be parallelized across chunks; each one
		// is fully independent because boundary vertices are held fixed.
		for ci := range chunks {
			chunks[ci] = quadricErrorMetricProcessChunk(chunks[ci])
		}
		lodVerts, lodIndices := stitchQemChunks(verts, chunks)
		if len(lodVerts) == 0 || len(lodIndices) == 0 {
			lods.Levels[i] = MeshLODInstance{Mesh: mesh, Ratio: ratio}
			continue
		}
		lodMesh := NewMesh(fmt.Sprintf("%s_lod_%d", mesh.Key(), i), lodVerts, lodIndices)
		lods.Levels[i] = MeshLODInstance{Mesh: lodMesh, Ratio: ratio}
	}
	// Remember the computed LODs so we don't redo them next call.
	mesh.lods = lods
	return lods, nil
}

// stitchQemChunks welds the per-chunk simplified results back into a single
// mesh. Vertices are de-duplicated by their original global index, so vertices
// shared across chunk boundaries (the fixed border vertices) map to a single
// output vertex and the result stays watertight. The collapsed (optimal)
// position from each representative is used.
func stitchQemChunks(verts []Vertex, chunks []MeshQemChunk) ([]Vertex, []uint32) {
	globalToNew := make(map[int]int, len(verts))
	outVerts := make([]Vertex, 0, len(verts))
	outIndices := make([]uint32, 0, len(verts))
	for ci := range chunks {
		chunk := &chunks[ci]
		for _, li := range chunk.resultIndices {
			local := int(li)
			g := chunk.globalIndices[local]
			newIdx, ok := globalToNew[g]
			if !ok {
				newIdx = len(outVerts)
				globalToNew[g] = newIdx
				v := verts[g]
				// Use the optimal collapse position from this chunk's result.
				v.Position = chunk.positions[local]
				outVerts = append(outVerts, v)
			}
			outIndices = append(outIndices, uint32(newIdx))
		}
	}
	return outVerts, outIndices
}

func (h qemEdgeHeap) Len() int           { return len(h) }
func (h qemEdgeHeap) Less(i, j int) bool { return h[i].cost < h[j].cost }
func (h qemEdgeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *qemEdgeHeap) Push(x any)        { *h = append(*h, x.(*qemEdge)) }
func (h *qemEdgeHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// quadricErrorMetricChunkify splits a mesh (vertices + indices) into a slice of
// spatially-local MeshQemChunks for future parallel QEM simplification. Chunking is
// performed at runtime (rather than offline) to support player mods that ship
// un-simplified geometry and are loaded on demand.
//
// Triangles are walked in index order and added to a group until the local
// vertex count reaches meshQemChunkTargetVerts, at which point the group is
// emitted as a chunk and a fresh group is started. A vertex that spans more
// than one group is duplicated into each chunk it belongs to and flagged as
// boundary, keeping the shared seams watertight. Those boundary vertices must
// be held fixed by processChunk.
//
// The ratio parameter (0, 1] is the target fraction of triangles to keep, so
// each chunk's targetTriangles is derived from ratio and its own triangle
// count.
//
// The returned slice is flat; the caller is expected to run processChunk over
// each chunk (e.g. in parallel) and stitch the per-chunk resultIndices back
// together.
func quadricErrorMetricChunkify(verts []Vertex, indices []uint32, ratio float32) []MeshQemChunk {
	if len(indices) == 0 || len(verts) == 0 || ratio <= 0 {
		return nil
	}
	if ratio > 1 {
		ratio = 1
	}
	// Maps a global vertex index to the (chunk index, local index) pairs it
	// occupies. A boundary vertex appears in more than one chunk.
	vertexSlots := make([][]meshQemSlot, len(verts))
	chunks := make([]MeshQemChunk, 0, 8)
	current := MeshQemChunk{}
	localToGlobal := make([]int, 0, meshQemChunkTargetVerts)
	globalToLocal := make(map[int]int, meshQemChunkTargetVerts)
	tris := len(indices) / 3
	// cornerLocal holds the remapped local index of each corner of the triangle
	// currently being processed.
	cornerLocal := [3]int{}
	addTriangle := func(base int) {
		for c := 0; c < 3; c++ {
			g := int(indices[base+c])
			if li, ok := globalToLocal[g]; ok {
				cornerLocal[c] = li
				continue
			}
			li := len(localToGlobal)
			localToGlobal = append(localToGlobal, g)
			globalToLocal[g] = li
			cornerLocal[c] = li
		}
		current.indices = append(current.indices,
			uint32(cornerLocal[0]), uint32(cornerLocal[1]), uint32(cornerLocal[2]))
	}
	for t := 0; t < tris; t++ {
		base := t * 3
		addTriangle(base)
		// If the current chunk overflowed its budget, emit it and restart the
		// group with the same triangle.
		if len(localToGlobal) > meshQemChunkTargetVerts {
			emitQemChunk(&chunks, &current, localToGlobal, vertexSlots, ratio)
			current = MeshQemChunk{}
			clear(globalToLocal)
			localToGlobal = localToGlobal[:0]
			current.indices = current.indices[:0]
			addTriangle(base)
		}
	}
	// Emit the final, possibly incomplete, chunk.
	if len(localToGlobal) > 0 {
		emitQemChunk(&chunks, &current, localToGlobal, vertexSlots, ratio)
	}
	buildQemChunkGeometry(verts, chunks, vertexSlots)
	return chunks
}

// emitQemChunk finalizes a collected chunk: it records each member vertex in
// vertexSlots (so boundary membership can be resolved later) and computes the
// chunk's targetTriangles from its own triangle count and the requested ratio.
func emitQemChunk(chunks *[]MeshQemChunk, chunk *MeshQemChunk, localToGlobal []int, vertexSlots [][]meshQemSlot, ratio float32) {
	chunkIndex := len(*chunks)
	for li, g := range localToGlobal {
		vertexSlots[g] = append(vertexSlots[g], meshQemSlot{chunk: chunkIndex, local: li})
	}
	chunkTriCount := len(chunk.indices) / 3
	if ratio >= 1 {
		chunk.targetTriangles = chunkTriCount
	} else {
		chunk.targetTriangles = int(float32(chunkTriCount)*ratio) + 1
		if chunk.targetTriangles > chunkTriCount {
			chunk.targetTriangles = chunkTriCount
		}
	}
	*chunks = append(*chunks, *chunk)
}

// buildQemChunkGeometry fills in each chunk's local position list and boundary
// flags. A vertex is on the boundary iff it appears in more than one chunk.
func buildQemChunkGeometry(verts []Vertex, chunks []MeshQemChunk, vertexSlots [][]meshQemSlot) {
	// Determine the size of each chunk (max local index + 1) so the position
	// and boundary arrays are allocated large enough before indexing them.
	chunkSizes := make([]int, len(chunks))
	for _, slots := range vertexSlots {
		for _, s := range slots {
			if s.local+1 > chunkSizes[s.chunk] {
				chunkSizes[s.chunk] = s.local + 1
			}
		}
	}
	for ci := range chunks {
		chunks[ci].positions = make([]matrix.Vec3, chunkSizes[ci])
		chunks[ci].boundary = make([]bool, chunkSizes[ci])
		chunks[ci].globalIndices = make([]int, chunkSizes[ci])
	}
	// First pass: populate positions (and the original global index) from the
	// source vertices.
	for g, slots := range vertexSlots {
		for _, s := range slots {
			chunks[s.chunk].positions[s.local] = verts[g].Position
			chunks[s.chunk].globalIndices[s.local] = g
		}
	}
	// Second pass: boundary flags = vertex shared by more than one chunk.
	for _, slots := range vertexSlots {
		if len(slots) <= 1 {
			continue
		}
		for _, s := range slots {
			chunks[s.chunk].boundary[s.local] = true
		}
	}
}

// selectMeshLodRatios returns the LOD screen-size ratios for the given number of levels.
// The ratios follow a power-of-two falloff: the first level is always 1.0 (100%) and each
// subsequent level is halved, so level i has ratio 1.0 / 2^i (e.g. 1.0, 0.5, 0.25, ...).
// This provides progressively coarser meshes as the object gets smaller on screen.
// The count 0 or negative is treated as no levels, returning an empty slice.
func selectMeshLodRatios(count int) []float32 {
	if count <= 0 {
		return []float32{}
	}
	ratios := make([]float32, count)
	for i := 0; i < count; i++ {
		ratios[i] = float32(1.0) / float32(uint32(1)<<uint32(i))
	}
	return ratios
}

func quadricErrorMetricProcessChunk(chunk MeshQemChunk) MeshQemChunk {
	// Degenerate or trivially small chunks pass through unchanged.
	tris := len(chunk.indices) / 3
	if tris < 1 {
		chunk.resultIndices = chunk.indices
		return chunk
	}
	if chunk.targetTriangles >= tris {
		chunk.resultIndices = chunk.indices
		return chunk
	}
	nVerts := len(chunk.positions)
	// A working copy of the positions; surviving vertices keep their index.
	positions := make([]matrix.Vec3, nVerts)
	copy(positions, chunk.positions)
	// Per-vertex quadric matrices; index i matches positions[i].
	quadrics := make([]matrix.Mat4, nVerts)
	// Live triangles, indexed by their first index into chunk.indices (3*t).
	liveTris := make([]bool, tris)
	// Union-find "parent": parent[i] is the representative vertex that vertex i
	// has collapsed into. find(i) resolves the current representative.
	parent := make([]int, nVerts)
	for i := range parent {
		parent[i] = i
	}
	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	// Vertex-to-triangle adjacency (built once, then maintained incrementally).
	adj := make([][]int, nVerts)
	// Build per-vertex quadrics and adjacency from the initial triangles.
	for t := 0; t < tris; t++ {
		a := int(chunk.indices[t*3+0])
		b := int(chunk.indices[t*3+1])
		c := int(chunk.indices[t*3+2])
		if a == b || b == c || a == c {
			liveTris[t] = false
			continue
		}
		liveTris[t] = true
		addPlaneQuadric(&quadrics[a], positions[a], positions[b], positions[c])
		addPlaneQuadric(&quadrics[b], positions[a], positions[b], positions[c])
		addPlaneQuadric(&quadrics[c], positions[a], positions[b], positions[c])
		adj[a] = append(adj[a], t)
		adj[b] = append(adj[b], t)
		adj[c] = append(adj[c], t)
	}
	// Build the initial set of candidate edges from the live triangles.
	edges := make(map[[2]int]bool)
	for t := 0; t < tris; t++ {
		if !liveTris[t] {
			continue
		}
		a := int(chunk.indices[t*3+0])
		b := int(chunk.indices[t*3+1])
		c := int(chunk.indices[t*3+2])
		addEdgeKey(edges, a, b)
		addEdgeKey(edges, b, c)
		addEdgeKey(edges, c, a)
	}
	// Priority queue of candidate edge collapses ordered by lowest error.
	pq := &qemEdgeHeap{}
	heap.Init(pq)
	edgeIndex := make(map[[2]int]*qemEdge)
	// addCandidate computes and enqueues a collapse candidate for the given
	// representative edge. The lower-index endpoint is treated as v0 (the one
	// that gets absorbed) whenever possible.
	addCandidate := func(k [2]int) {
		if _, exists := edgeIndex[k]; exists {
			return
		}
		r0, r1 := find(k[0]), find(k[1])
		if r0 == r1 {
			return
		}
		// Only a non-boundary vertex may be collapsed away. Choose the absorbed
		// vertex to be the one that is allowed to move (not a boundary vertex).
		// If both are boundary, there's nothing to collapse.
		v0, v1 := r0, r1
		if chunk.boundary[v0] && !chunk.boundary[v1] {
			v0, v1 = v1, v0
		} else if chunk.boundary[v0] && chunk.boundary[v1] {
			return
		}
		combined := quadrics[v0]
		combined.AddAssign(quadrics[v1])
		cost, pos, ok := solveQuadric(combined)
		if !ok {
			cost = quadricCost(combined, positions[v0])
			pos = positions[v0]
		}
		e := &qemEdge{v0: v0, v1: v1, cost: cost, pos: pos}
		e.key = [2]int{v0, v1}
		edgeIndex[e.key] = e
		heap.Push(pq, e)
	}
	for k := range edges {
		addCandidate(k)
	}
	liveCount := tris
	for pq.Len() > 0 {
		if liveCount <= chunk.targetTriangles {
			break
		}
		e := heap.Pop(pq).(*qemEdge)
		// Resolve current representatives; stale entries (already collapsed) are
		// dropped.
		r0, r1 := find(e.v0), find(e.v1)
		if r0 == r1 {
			continue
		}
		if chunk.boundary[e.v0] && !chunk.boundary[e.v1] {
			// Re-queue with the roles swapped so the movable vertex is absorbed.
			addCandidate([2]int{e.v1, e.v0})
			continue
		}
		// Perform the collapse into representative r1 (the surviving vertex).
		merged, removed := applyCollapse(e, parent, positions, quadrics, adj, liveTris, chunk.boundary)
		if !merged {
			continue
		}
		_ = r0
		liveCount -= removed
		// Refresh candidates around the merged vertex.
		refreshCandidates(edges, pq, edgeIndex, parent, chunk.boundary, addCandidate)
	}
	// Emit the surviving triangles, resolving each corner to its representative.
	chunk.resultIndices = chunk.resultIndices[:0]
	for t := 0; t < tris; t++ {
		if !liveTris[t] {
			continue
		}
		chunk.resultIndices = append(chunk.resultIndices,
			uint32(find(int(chunk.indices[t*3+0]))),
			uint32(find(int(chunk.indices[t*3+1]))),
			uint32(find(int(chunk.indices[t*3+2]))))
	}
	// Write back the collapsed (representative) positions so the stitch step can
	// produce a final mesh with the optimal collapse positions.
	copy(chunk.positions, positions)
	chunk.quadrics = quadrics
	return chunk
}

// addPlaneQuadric accumulates the quadric of the plane through the triangle
// (p0, p1, p2) into *q. The plane is [n, d] where n is the (unnormalized) cross
// product normal and d = -n·p0, giving plane equation n·x + d = 0. The error of
// a point x is (n·x + d)^2, so the quadric is K = [A, b; b^T, c] with
// A = n·n^T, b = n*d, c = d*d. Skipping the normalization is the standard
// simplification that preserves the squared-distance weighting.
func addPlaneQuadric(q *matrix.Mat4, p0, p1, p2 matrix.Vec3) {
	n := p1.Subtract(p0).Cross(p2.Subtract(p0))
	d := -n.Dot(p0)
	// A = n n^T (symmetric 3x3, upper triangle) + b + c.
	nx, ny, nz := n.X(), n.Y(), n.Z()
	q[0] += nx * nx
	q[1] += nx * ny
	q[2] += nx * nz
	q[3] += nx * d
	q[5] += ny * ny
	q[6] += ny * nz
	q[7] += ny * d
	q[10] += nz * nz
	q[11] += nz * d
	q[15] += d * d
	// Mirror the symmetric upper triangle into the lower half.
	q[4] = q[1]
	q[8] = q[2]
	q[9] = q[6]
	q[12] = q[3]
	q[13] = q[7]
	q[14] = q[11]
}

// quadricCost evaluates v^T Q v for a 4x4 symmetric quadric stored with the
// 3x3 block in indices {0,1,2,5,6,10}, the affine part b in {3,7,11} and the
// constant c in 15 (see addPlaneQuadric).
func quadricCost(q matrix.Mat4, v matrix.Vec3) matrix.Float {
	x, y, z := v.X(), v.Y(), v.Z()
	return x*(q[0]*x+q[1]*y+q[2]*z+q[3]) +
		y*(q[1]*x+q[5]*y+q[6]*z+q[7]) +
		z*(q[2]*x+q[6]*y+q[10]*z+q[11]) + q[15]
}

// solveQuadric finds the point v minimizing v^T Q v by solving the 3x3 linear
// system A v = -b, where A is the 3x3 block of the quadric and b is the affine
// part. Returns ok=false when A is singular (no unique minimum); the caller
// should fall back to one of the endpoints.
func solveQuadric(q matrix.Mat4) (matrix.Float, matrix.Vec3, bool) {
	a00, a01, a02 := q[0], q[1], q[2]
	a10, a11, a12 := q[1], q[5], q[6]
	a20, a21, a22 := q[2], q[6], q[10]
	b0, b1, b2 := q[3], q[7], q[11]
	det := a00*(a11*a22-a12*a21) - a01*(a10*a22-a12*a20) + a02*(a10*a21-a11*a20)
	eps := matrix.Float(1e-12)
	if det > -eps && det < eps {
		return 0, matrix.Vec3{}, false
	}
	inv := 1.0 / det
	// Inverse of A via cofactors: inv[i][j] = Cji / det.
	c00 := a11*a22 - a12*a21
	c01 := -(a10*a22 - a12*a20)
	c02 := a10*a21 - a11*a20
	c10 := -(a01*a22 - a02*a21)
	c11 := a00*a22 - a02*a20
	c12 := -(a00*a21 - a01*a20)
	c20 := a01*a12 - a02*a11
	c21 := -(a00*a12 - a02*a10)
	c22 := a00*a11 - a01*a10
	// x = A^-1 * (-b); inv[i][j] = Cji.
	x := (c00*-b0 + c10*-b1 + c20*-b2) * inv
	y := (c01*-b0 + c11*-b1 + c21*-b2) * inv
	z := (c02*-b0 + c12*-b1 + c22*-b2) * inv
	pos := matrix.Vec3{x, y, z}
	return quadricCost(q, pos), pos, true
}

// applyCollapse merges the representative of e.v0 into e.v1. Every triangle
// that becomes degenerate (references both endpoints) is removed, and all other
// triangles touching the absorbed vertex are rewired via the union-find parent
// so they now reference the representative. It returns whether the collapse was
// performed and how many triangles were removed.
func applyCollapse(e *qemEdge, parent []int, positions []matrix.Vec3, quadrics []matrix.Mat4, adj [][]int, liveTris, boundary []bool) (bool, int) {
	// e.v0 must be a non-boundary vertex being absorbed into e.v1.
	v0, v1 := e.v0, e.v1
	if boundary[v0] {
		return false, 0
	}
	removed := 0
	for _, t := range adj[v0] {
		if !liveTris[t] {
			continue
		}
		// Determine whether this triangle also references v1. If so it becomes
		// degenerate after the merge and is removed.
		if triUses(adj, v1, t) {
			liveTris[t] = false
			removed++
			continue
		}
		// Otherwise the triangle survives; the union-find rewrite at emit time
		// maps v0 -> v1. We leave liveTris as is.
		_ = t
	}
	// Union: v0's parent becomes v1.
	parent[v0] = v1
	// v1 inherits v0's quadric and moves to the optimal position.
	quadrics[v1].AddAssign(quadrics[v0])
	positions[v1] = e.pos
	return true, removed
}

// triUses reports whether any live triangle touching vertex v is t.
func triUses(adj [][]int, v, t int) bool {
	for _, x := range adj[v] {
		if x == t {
			return true
		}
	}
	return false
}

// addEdgeKey records the undirected edge (a, b) into the set.
func addEdgeKey(edges map[[2]int]bool, a, b int) {
	if a > b {
		a, b = b, a
	}
	edges[[2]int{a, b}] = true
}

// refreshCandidates rebuilds the candidate queue after a collapse: it drops
// stale entries whose endpoints now share a representative and re-adds the
// edges surrounding the merged vertex.
func refreshCandidates(edges map[[2]int]bool, pq *qemEdgeHeap, edgeIndex map[[2]int]*qemEdge, parent []int, boundary []bool, addCandidate func(k [2]int)) {
	// Drop candidates whose endpoints now resolve to the same representative.
	for k, e := range edgeIndex {
		if findParent(parent, e.v0) == findParent(parent, e.v1) {
			delete(edgeIndex, k)
		}
	}
	// Re-add the current edge set (stale candidates get filtered inside
	// addCandidate by representative equality).
	for k := range edges {
		addCandidate(k)
	}
}

// findParent resolves the representative of x in the union-find parent array.
func findParent(parent []int, x int) int {
	for parent[x] != x {
		x = parent[x]
	}
	return x
}

// countLiveTris counts how many triangles are still alive.
func countLiveTris(liveTris []bool) int {
	n := 0
	for _, l := range liveTris {
		if l {
			n++
		}
	}
	return n
}

// qemEdge is a candidate edge collapse in the priority queue.
type qemEdge struct {
	v0, v1 int
	cost   matrix.Float
	pos    matrix.Vec3
	key    [2]int
}
