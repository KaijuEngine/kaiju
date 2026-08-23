package rendering

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"sync"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

const (
	// meshQemChunkTargetVerts is the maximum number of vertices a single chunk may
	// hold before the mesh is split again. It bounds per-goroutine work so chunk
	// processing stays short enough to run at load/runtime without a perceptible
	// hitch.
	meshQemChunkTargetVerts = 4096

	// meshLodMinTriangles is the triangle-count floor below which a mesh is
	// considered "already low poly enough" to skip further simplification.
	// Meshes at or below this count (e.g. cubes = 12 tris, quads = 2 tris, and
	// other small primitives) reuse the source mesh for every LOD level, and a
	// generated LOD that reaches this floor becomes the final LOD for all
	// remaining levels.
	meshLodMinTriangles = 32
)

type MeshLodGeneratorQem struct {
	Threads *concurrent.Threads
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
// vertex is placed at the lowest-error point along the source edge, and its
// quadric is replaced by the sum of the two collapsed vertices' quadrics so the
// error estimate is carried forward. Restricting the solution to the edge and
// rejecting topology or normal inversions keeps the runtime-generated mesh
// valid while retaining the greedy, error-driven behavior of QEM.

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
//   - normals:         the corresponding source normals, used to preserve face
//     orientation while validating collapses.
//   - indices:         the chunk's triangles, remapped into local position space.
//   - sourceIndices:   the original attribute vertex for each triangle corner.
//   - boundary:        for each local position, whether it lies on the chunk's
//     outer border. Border vertices are shared with neighboring chunks and MUST
//     be held fixed during simplification so adjacent chunks stay watertight
//     and no seams appear when the results are stitched back together.
//   - globalIndices:   for each local position, its welded geometry index in
//     the full mesh. Used to weld chunks back together.
//   - targetTriangles: the triangle budget this chunk should collapse down to,
//     derived from the requested simplification ratio.
//
// Outputs populated by processChunk:
//   - quadrics:      the per-vertex QEM matrices (len == len(positions)).
//   - resultIndices: the simplified triangle list, referencing local position
//     indices, ready to be merged with the other chunks' results.
//   - resultSources: the original attribute vertex for each result corner.
type MeshQemChunk struct {
	// Inputs
	positions       []matrix.Vec3
	normals         []matrix.Vec3
	indices         []uint32
	sourceIndices   []int
	sourceVariants  [][]int
	boundary        []bool
	globalIndices   []int
	targetTriangles int

	// Outputs
	quadrics      []matrix.Mat4
	resultIndices []uint32
	resultSources []int
}

// meshQemWeldData separates geometric vertices from attribute vertices. OBJ,
// FBX, glTF, and generated UV meshes commonly duplicate a position at UV or
// hard-normal seams. QEM operates on the welded geometry while sourceVariants
// retain every original vertex needed to reconstruct those attributes later.
type meshQemWeldData struct {
	sourceToWeld []int
	positions    []matrix.Vec3
	normals      []matrix.Vec3
	variants     [][]int
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

func NewMeshLodGeneratorQem(threads *concurrent.Threads) *MeshLodGeneratorQem {
	return &MeshLodGeneratorQem{
		Threads: threads,
	}
}

// generateMeshLOD builds a set of level-of-detail meshes for the given source
// mesh using Quadric Error Metric (QEM) simplification.
//
// The result is a MeshLod whose Level 0 is always the full-resolution source
// mesh, followed by levels 1..levels-1 whose triangle counts fall off
// power-of-two (see selectMeshLodRatios): level i has ratio 1.0/2^i.
//
// Generation happens at runtime (rather than being pre-baked offline) so that
// content loaded from player mods can be simplified on demand. The algorithm
// chunks the mesh (quadricErrorMetricChunkify), simplifies each chunk
// independently in parallel (quadricErrorMetricProcessChunk — safe because
// boundary vertices are held fixed), and welds the chunk results back into a
// single watertight mesh (stitchQemChunks).
//
// The computed LODs are cached on mesh so repeated calls return the same set.
// verts and indices describe the full-resolution geometry; if a level cannot
// be reduced, that level falls back to reusing the source mesh itself.
func (g *MeshLodGeneratorQem) GenerateLods(mesh *Mesh, cache *MeshCache, verts []Vertex, indices []uint32, levels int) (MeshLod, error) {
	if levels <= 0 {
		return MeshLod{}, errors.New("generateMeshLOD: levels must be greater than 0")
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
	lastLod := mesh
	// If the source mesh is already low poly enough, there is nothing to
	// simplify. Every level reuses the source mesh itself and we skip the
	// (pointless) chunking/simplification work entirely.
	if len(indices)/3 <= meshLodMinTriangles {
		for i := 1; i < levels; i++ {
			lods.Levels[i] = MeshLODInstance{Mesh: mesh, Ratio: ratios[i]}
		}
		mesh.lods = lods
		return lods, nil
	}
	for i := 1; i < levels; i++ {
		ratio := ratios[i]
		chunks := quadricErrorMetricChunkify(verts, indices, ratio)
		if len(chunks) == 0 {
			// Nothing to simplify; reuse the full mesh for this level.
			lods.Levels[i] = MeshLODInstance{Mesh: lastLod, Ratio: ratio}
			continue
		}
		// Chunks are independent after geometric seam welding and boundary
		// locking, so simplification can be distributed across the generator's
		// worker threads.
		g.processQemChunks(chunks)
		lodVerts, lodIndices := stitchQemChunks(verts, chunks)
		if len(lodVerts) == 0 || len(lodIndices) == 0 {
			lods.Levels[i] = MeshLODInstance{Mesh: lastLod, Ratio: ratio}
			continue
		}
		lastLod = cache.meshLod(fmt.Sprintf("%s_lod_%d", mesh.Key(), i), lodVerts, lodIndices)
		lods.Levels[i] = MeshLODInstance{Mesh: lastLod, Ratio: ratio}
		// Once a generated LOD reaches the low-poly floor, it is the coarsest
		// useful representation. Make every remaining level reuse this same LOD
		// rather than continuing to simplify.
		if len(lodIndices)/3 <= meshLodMinTriangles {
			for j := i + 1; j < levels; j++ {
				lods.Levels[j] = MeshLODInstance{Mesh: lastLod, Ratio: ratios[j]}
			}
			break
		}
	}
	// Remember the computed LODs so we don't redo them next call.
	mesh.lods = lods
	return lods, nil
}

// processQemChunks simplifies every chunk, dispatching the work to the
// generator's worker threads when they are available. Each chunk is processed
// independently (boundary vertices are held fixed), so they can run in parallel
// without sharing state. When no worker threads are configured, the chunks are
// processed inline.
func (g *MeshLodGeneratorQem) processQemChunks(chunks []MeshQemChunk) {
	threads := g.Threads
	if threads == nil || threads.ThreadCount() == 0 || len(chunks) <= 1 {
		for i := range chunks {
			chunks[i] = quadricErrorMetricProcessChunk(chunks[i])
		}
		return
	}
	group := sync.WaitGroup{}
	group.Add(len(chunks))
	work := make([]func(int), len(chunks))
	for i := range chunks {
		// Capture i to avoid the classic loop-variable aliasing bug.
		i := i
		work[i] = func(int) {
			chunks[i] = quadricErrorMetricProcessChunk(chunks[i])
			group.Done()
		}
	}
	threads.AddWork(work)
	group.Wait()
}

// stitchQemChunks welds the per-chunk simplified results back into a single
// mesh. Geometry shared across chunk boundaries uses the same fixed position,
// while the source+geometry key intentionally recreates duplicate vertices for
// UV, normal, tangent, or color seams. The collapsed position from each
// geometric representative is applied to every attribute variant.
func stitchQemChunks(verts []Vertex, chunks []MeshQemChunk) ([]Vertex, []uint32) {
	type stitchVertexKey struct {
		source   int
		geometry int
	}
	globalToNew := make(map[stitchVertexKey]int, len(verts))
	outVerts := make([]Vertex, 0, len(verts))
	outIndices := make([]uint32, 0, len(verts))
	for ci := range chunks {
		chunk := &chunks[ci]
		for corner, li := range chunk.resultIndices {
			local := int(li)
			source := -1
			if corner < len(chunk.resultSources) {
				source = chunk.resultSources[corner]
			}
			if (source < 0 || source >= len(verts)) && local < len(chunk.sourceVariants) && len(chunk.sourceVariants[local]) > 0 {
				source = chunk.sourceVariants[local][0]
			}
			if source < 0 || source >= len(verts) {
				// Chunkified production data always has a source variant. Returning
				// no mesh is safer than emitting an index list with a missing corner.
				return nil, nil
			}
			key := stitchVertexKey{source: source, geometry: chunk.globalIndices[local]}
			newIdx, ok := globalToNew[key]
			if !ok {
				newIdx = len(outVerts)
				globalToNew[key] = newIdx
				v := verts[source]
				// Use the optimal collapse position from this chunk's result.
				v.Position = chunk.positions[local]
				outVerts = append(outVerts, v)
			}
			outIndices = append(outIndices, uint32(newIdx))
		}
	}
	return outVerts, outIndices
}

func (h qemEdgeHeap) Len() int { return len(h) }
func (h qemEdgeHeap) Less(i, j int) bool {
	if h[i].cost != h[j].cost {
		return h[i].cost < h[j].cost
	}
	if h[i].v0 != h[j].v0 {
		return h[i].v0 < h[j].v0
	}
	return h[i].v1 < h[j].v1
}
func (h qemEdgeHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *qemEdgeHeap) Push(x any)   { *h = append(*h, x.(*qemEdge)) }
func (h *qemEdgeHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

// buildMeshQemWeldData groups positions that differ only by floating-point seam
// noise. The tolerance is relative to the mesh diagonal and deliberately small:
// it joins duplicated attribute vertices without merging visibly distinct
// geometry.
func buildMeshQemWeldData(verts []Vertex) meshQemWeldData {
	result := meshQemWeldData{sourceToWeld: make([]int, len(verts))}
	if len(verts) == 0 {
		return result
	}
	low, high := verts[0].Position, verts[0].Position
	for i := 1; i < len(verts); i++ {
		low = matrix.Vec3Min(low, verts[i].Position)
		high = matrix.Vec3Max(high, verts[i].Position)
	}
	tolerance := float64(high.Subtract(low).Length()) * 1e-6
	if tolerance == 0 {
		tolerance = 1e-6
	}
	toleranceSquared := matrix.Float(tolerance * tolerance)
	type cellKey [3]int64
	cellFor := func(p matrix.Vec3) cellKey {
		return cellKey{
			int64(math.Floor(float64(p.X()) / tolerance)),
			int64(math.Floor(float64(p.Y()) / tolerance)),
			int64(math.Floor(float64(p.Z()) / tolerance)),
		}
	}
	cells := make(map[cellKey][]int, len(verts))
	for source := range verts {
		position := verts[source].Position
		cell := cellFor(position)
		weld := -1
		for z := int64(-1); z <= 1 && weld < 0; z++ {
			for y := int64(-1); y <= 1 && weld < 0; y++ {
				for x := int64(-1); x <= 1 && weld < 0; x++ {
					for _, candidate := range cells[cellKey{cell[0] + x, cell[1] + y, cell[2] + z}] {
						representative := result.variants[candidate][0]
						if position.SquareDistance(result.positions[candidate]) <= toleranceSquared &&
							qemVerticesDeformTogether(verts[source], verts[representative]) {
							weld = candidate
							break
						}
					}
				}
			}
		}
		if weld < 0 {
			weld = len(result.positions)
			result.positions = append(result.positions, position)
			result.normals = append(result.normals, matrix.Vec3Zero())
			result.variants = append(result.variants, nil)
			cells[cell] = append(cells[cell], weld)
		}
		result.sourceToWeld[source] = weld
		result.normals[weld] = result.normals[weld].Add(verts[source].Normal)
		result.variants[weld] = append(result.variants[weld], source)
	}
	for i := range result.normals {
		if !result.normals[i].IsZero() {
			result.normals[i].Normalize()
		}
	}
	return result
}

// qemVerticesDeformTogether prevents a geometric weld from joining vertices
// that only coincide in the bind pose. UVs, normals, tangents, and colors may
// differ across an attribute seam, but skinning and morph data must agree or
// the reconstructed duplicates could separate again while animating.
func qemVerticesDeformTogether(a, b Vertex) bool {
	return a.JointIds == b.JointIds &&
		a.JointWeights == b.JointWeights &&
		a.MorphTarget == b.MorphTarget
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
// Before partitioning, position-equivalent attribute vertices are welded into
// one geometric topology. This closes UV/hard-normal seams for simplification;
// sourceIndices retain the original per-corner attributes for reconstruction.
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
	welded := buildMeshQemWeldData(verts)
	// Maps a global welded vertex index to the (chunk index, local index) pairs it
	// occupies. A boundary vertex appears in more than one chunk.
	vertexSlots := make([][]meshQemSlot, len(welded.positions))
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
			source := int(indices[base+c])
			g := welded.sourceToWeld[source]
			if li, ok := globalToLocal[g]; ok {
				cornerLocal[c] = li
			} else {
				li := len(localToGlobal)
				localToGlobal = append(localToGlobal, g)
				globalToLocal[g] = li
				cornerLocal[c] = li
			}
			current.sourceIndices = append(current.sourceIndices, source)
		}
		current.indices = append(current.indices,
			uint32(cornerLocal[0]), uint32(cornerLocal[1]), uint32(cornerLocal[2]))
	}
	for t := 0; t < tris; t++ {
		base := t * 3
		newVerts := 0
		triangleWelds := [3]int{-1, -1, -1}
		for c := 0; c < 3; c++ {
			g := welded.sourceToWeld[int(indices[base+c])]
			alreadyCounted := false
			for p := 0; p < c; p++ {
				if triangleWelds[p] == g {
					alreadyCounted = true
					break
				}
			}
			triangleWelds[c] = g
			if _, ok := globalToLocal[g]; !ok && !alreadyCounted {
				newVerts++
			}
		}
		// Emit before adding the triangle that would exceed the chunk budget.
		// Adding it first and then restarting with the same triangle duplicates
		// that triangle in both chunks when the chunks are stitched together.
		if len(current.indices) > 0 && len(localToGlobal)+newVerts > meshQemChunkTargetVerts {
			emitQemChunk(&chunks, &current, localToGlobal, vertexSlots, ratio)
			current = MeshQemChunk{}
			clear(globalToLocal)
			localToGlobal = localToGlobal[:0]
		}
		addTriangle(base)
	}
	// Emit the final, possibly incomplete, chunk.
	if len(localToGlobal) > 0 {
		emitQemChunk(&chunks, &current, localToGlobal, vertexSlots, ratio)
	}
	buildQemChunkGeometry(welded, chunks, vertexSlots)
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

// buildQemChunkGeometry fills in each chunk's local welded positions, source
// variants, and boundary flags. A vertex is on the boundary iff it appears in
// more than one chunk.
func buildQemChunkGeometry(welded meshQemWeldData, chunks []MeshQemChunk, vertexSlots [][]meshQemSlot) {
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
		chunks[ci].normals = make([]matrix.Vec3, chunkSizes[ci])
		chunks[ci].sourceVariants = make([][]int, chunkSizes[ci])
		chunks[ci].boundary = make([]bool, chunkSizes[ci])
		chunks[ci].globalIndices = make([]int, chunkSizes[ci])
	}
	// First pass: populate positions (and the original global index) from the
	// source vertices.
	for g, slots := range vertexSlots {
		for _, s := range slots {
			chunks[s.chunk].positions[s.local] = welded.positions[g]
			chunks[s.chunk].normals[s.local] = welded.normals[g]
			chunks[s.chunk].sourceVariants[s.local] = welded.variants[g]
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
	tris := len(chunk.indices) / 3
	if tris < 1 || len(chunk.positions) == 0 {
		chunk.resultIndices = chunk.indices
		return chunk
	}
	positions := append([]matrix.Vec3(nil), chunk.positions...)
	normals := make([]matrix.Vec3, len(positions))
	copy(normals, chunk.normals)
	triangles := make([][3]int, tris)
	triangleSources := make([][3]int, tris)
	liveTris := make([]bool, tris)
	referenceFaces := make([]matrix.Vec3, tris)
	quadrics := make([]matrix.Mat4, len(positions))
	areaEpsilon := qemAreaEpsilon(positions)
	liveCount := 0
	for t := 0; t < tris; t++ {
		a := int(chunk.indices[t*3])
		b := int(chunk.indices[t*3+1])
		c := int(chunk.indices[t*3+2])
		triangles[t] = [3]int{a, b, c}
		for corner, local := range triangles[t] {
			source := -1
			inputCorner := t*3 + corner
			if inputCorner < len(chunk.sourceIndices) {
				source = chunk.sourceIndices[inputCorner]
			} else if local >= 0 && local < len(chunk.sourceVariants) && len(chunk.sourceVariants[local]) > 0 {
				source = chunk.sourceVariants[local][0]
			}
			triangleSources[t][corner] = source
		}
		if a < 0 || b < 0 || c < 0 || a >= len(positions) || b >= len(positions) || c >= len(positions) ||
			a == b || b == c || a == c {
			continue
		}
		face := positions[b].Subtract(positions[a]).Cross(positions[c].Subtract(positions[a]))
		if face.LengthSquared() <= areaEpsilon {
			continue
		}
		referenceFaces[t] = face
		liveTris[t] = true
		liveCount++
		addPlaneQuadric(&quadrics[a], positions[a], positions[b], positions[c])
		addPlaneQuadric(&quadrics[b], positions[a], positions[b], positions[c])
		addPlaneQuadric(&quadrics[c], positions[a], positions[b], positions[c])
	}
	// Inter-chunk vertices remain fixed so independently processed chunks stitch
	// to the same position. Real topological boundary vertices may collapse only
	// along other boundary vertices; this preserves the boundary without freezing
	// the many duplicated seam and pole vertices used by UV meshes.
	protected := make([]bool, len(positions))
	copy(protected, chunk.boundary)
	topologyBoundary := make([]bool, len(positions))
	_, initialEdgeUses, _ := qemTopology(triangles, liveTris, len(positions))
	for edge, uses := range initialEdgeUses {
		if uses != 2 {
			topologyBoundary[edge[0]] = true
			topologyBoundary[edge[1]] = true
		}
	}
	for liveCount > chunk.targetTriangles {
		neighbors, edgeUses, edges := qemTopology(triangles, liveTris, len(positions))
		pq := &qemEdgeHeap{}
		heap.Init(pq)
		for _, edge := range edges {
			v0, v1 := edge[0], edge[1]
			if topologyBoundary[v0] != topologyBoundary[v1] {
				continue
			}
			if protected[v0] && protected[v1] {
				continue
			}
			// v0 is removed and v1 survives. A protected endpoint may survive,
			// but it must never be removed or moved.
			if protected[v0] {
				v0, v1 = v1, v0
			}
			combined := quadrics[v0]
			combined.AddAssign(quadrics[v1])
			cost, pos := qemCollapsePosition(combined, positions[v0], positions[v1], protected[v1])
			if matrix.IsNaN(cost) || matrix.IsInf(cost, 0) || pos.IsNaN() || pos.IsInf(0) {
				continue
			}
			heap.Push(pq, &qemEdge{v0: v0, v1: v1, cost: cost, pos: pos})
		}
		collapsed := false
		for pq.Len() > 0 {
			e := heap.Pop(pq).(*qemEdge)
			if !qemCollapsePreservesTopology(e.v0, e.v1, neighbors, edgeUses) ||
				!qemCollapsePreservesGeometry(e, triangles, liveTris, positions, normals, referenceFaces, areaEpsilon) {
				continue
			}
			removed := qemApplyCollapse(e, triangles, liveTris, positions, quadrics)
			if removed == 0 {
				continue
			}
			liveCount -= removed
			collapsed = true
			break
		}
		if !collapsed {
			break
		}
	}
	chunk.resultIndices = chunk.resultIndices[:0]
	chunk.resultSources = chunk.resultSources[:0]
	for t := range triangles {
		if !liveTris[t] {
			continue
		}
		chunk.resultIndices = append(chunk.resultIndices,
			uint32(triangles[t][0]), uint32(triangles[t][1]), uint32(triangles[t][2]))
		chunk.resultSources = append(chunk.resultSources,
			triangleSources[t][0], triangleSources[t][1], triangleSources[t][2])
	}
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

// qemAreaEpsilon returns a scale-relative threshold for rejecting zero-area
// faces. This also removes the numerically tiny cap triangles produced by UV
// spheres whose duplicated pole positions differ only by trig roundoff.
func qemAreaEpsilon(positions []matrix.Vec3) matrix.Float {
	if len(positions) == 0 {
		return 0
	}
	low, high := positions[0], positions[0]
	for i := 1; i < len(positions); i++ {
		low = matrix.Vec3Min(low, positions[i])
		high = matrix.Vec3Max(high, positions[i])
	}
	scaleSquared := high.Subtract(low).LengthSquared()
	return scaleSquared * scaleSquared * matrix.Float(1e-12)
}

// qemCollapsePosition selects the lowest-cost finite point along the collapsed
// edge. Constraining the optimum to the edge segment prevents a mathematically
// low-cost solution from jumping across a curved mesh. A protected survivor is
// held exactly at its source position so independently processed chunks and open
// seams remain watertight.
func qemCollapsePosition(q matrix.Mat4, removed, survivor matrix.Vec3, survivorProtected bool) (matrix.Float, matrix.Vec3) {
	if survivorProtected {
		return quadricCost(q, survivor), survivor
	}
	midpoint := removed.Add(survivor).Scale(0.5)
	candidates := [4]matrix.Vec3{
		removed,
		survivor,
		midpoint,
	}
	candidateCount := 3
	// Fit f(t) = a*t^2 + b*t + c from the costs at t={0,.5,1},
	// then include its segment-clamped minimum as the fourth candidate.
	f0 := quadricCost(q, removed)
	f1 := quadricCost(q, survivor)
	fm := quadricCost(q, midpoint)
	a := 2*f0 + 2*f1 - 4*fm
	b := f1 - f0 - a
	if matrix.Abs(a) > matrix.Float(1e-20) {
		t := -b / (2 * a)
		if t > 0 && t < 1 {
			candidates[candidateCount] = removed.Add(survivor.Subtract(removed).Scale(t))
			candidateCount++
		}
	}
	bestCost := matrix.FloatMax
	bestPosition := survivor
	for i := 0; i < candidateCount; i++ {
		cost := quadricCost(q, candidates[i])
		if matrix.IsNaN(cost) || matrix.IsInf(cost, 0) {
			continue
		}
		if cost < bestCost {
			bestCost = cost
			bestPosition = candidates[i]
		}
	}
	return bestCost, bestPosition
}

// qemTopology rebuilds the live one-ring neighborhoods and edge-use counts.
// Rebuilding after each accepted collapse ensures every queued edge references
// the current topology; stale endpoints were the source of the disappearing
// triangles in the previous implementation.
func qemTopology(triangles [][3]int, liveTris []bool, vertexCount int) ([]map[int]struct{}, map[[2]int]int, [][2]int) {
	neighbors := make([]map[int]struct{}, vertexCount)
	edgeUses := make(map[[2]int]int)
	add := func(a, b int) {
		if a > b {
			a, b = b, a
		}
		edgeUses[[2]int{a, b}]++
	}
	addNeighbor := func(a, b int) {
		if neighbors[a] == nil {
			neighbors[a] = make(map[int]struct{})
		}
		neighbors[a][b] = struct{}{}
	}
	for t, tri := range triangles {
		if !liveTris[t] {
			continue
		}
		a, b, c := tri[0], tri[1], tri[2]
		add(a, b)
		add(b, c)
		add(c, a)
		addNeighbor(a, b)
		addNeighbor(a, c)
		addNeighbor(b, a)
		addNeighbor(b, c)
		addNeighbor(c, a)
		addNeighbor(c, b)
	}
	edges := make([][2]int, 0, len(edgeUses))
	for edge := range edgeUses {
		edges = append(edges, edge)
	}
	return neighbors, edgeUses, edges
}

// qemCollapsePreservesTopology applies the triangle-mesh link condition. For a
// manifold edge, the endpoints may have only the opposite vertices of the
// edge's one or two incident triangles in common. Any additional shared neighbor
// would turn the collapse into a duplicate face or non-manifold connection.
func qemCollapsePreservesTopology(removed, survivor int, neighbors []map[int]struct{}, edgeUses map[[2]int]int) bool {
	a, b := removed, survivor
	if a > b {
		a, b = b, a
	}
	uses := edgeUses[[2]int{a, b}]
	if uses < 1 || uses > 2 {
		return false
	}
	common := 0
	for neighbor := range neighbors[removed] {
		if _, ok := neighbors[survivor][neighbor]; ok {
			common++
		}
	}
	return common == uses
}

// qemCollapsePreservesGeometry rejects collapses that flatten or flip any
// surviving incident face. Faces containing both endpoints are intentionally
// omitted because the collapse removes them.
func qemCollapsePreservesGeometry(e *qemEdge, triangles [][3]int, liveTris []bool, positions, normals, referenceFaces []matrix.Vec3, areaEpsilon matrix.Float) bool {
	for t, tri := range triangles {
		if !liveTris[t] {
			continue
		}
		usesRemoved := tri[0] == e.v0 || tri[1] == e.v0 || tri[2] == e.v0
		usesSurvivor := tri[0] == e.v1 || tri[1] == e.v1 || tri[2] == e.v1
		if !usesRemoved && !usesSurvivor {
			continue
		}
		if usesRemoved && usesSurvivor {
			continue
		}
		updated := [3]matrix.Vec3{positions[tri[0]], positions[tri[1]], positions[tri[2]]}
		updatedIndices := tri
		for i := range tri {
			if tri[i] == e.v0 || tri[i] == e.v1 {
				updated[i] = e.pos
			}
			if tri[i] == e.v0 {
				updatedIndices[i] = e.v1
			}
		}
		newFace := updated[1].Subtract(updated[0]).Cross(updated[2].Subtract(updated[0]))
		newArea := newFace.LengthSquared()
		referenceArea := referenceFaces[t].LengthSquared()
		direction := referenceFaces[t].Dot(newFace)
		// Require more than a merely positive dot product. Repeated collapses can
		// otherwise rotate a face almost 90 degrees a little at a time, producing
		// radial "curtains" that look like missing wedges under back-face culling.
		const minNormalDotSquared = matrix.Float(0.04) // cos(angle) >= 0.2
		if newFace.IsNaN() || newFace.IsInf(0) || newArea <= areaEpsilon || direction <= 0 ||
			direction*direction < referenceArea*newArea*minNormalDotSquared {
			return false
		}
		vertexNormal := normals[updatedIndices[0]].Add(normals[updatedIndices[1]]).Add(normals[updatedIndices[2]])
		if !vertexNormal.IsZero() {
			normalDirection := vertexNormal.Dot(newFace)
			if normalDirection <= 0 || normalDirection*normalDirection < vertexNormal.LengthSquared()*newArea*minNormalDotSquared {
				return false
			}
		}
	}
	return true
}

// qemApplyCollapse rewires the current triangle indices immediately, removing
// faces that contain both endpoints. No deferred union-find resolution is used,
// so subsequent adjacency and candidate costs are built from the actual mesh.
func qemApplyCollapse(e *qemEdge, triangles [][3]int, liveTris []bool, positions []matrix.Vec3, quadrics []matrix.Mat4) int {
	removedTriangles := 0
	for t := range triangles {
		if !liveTris[t] {
			continue
		}
		for i := range triangles[t] {
			if triangles[t][i] == e.v0 {
				triangles[t][i] = e.v1
			}
		}
		tri := triangles[t]
		if tri[0] == tri[1] || tri[1] == tri[2] || tri[0] == tri[2] {
			liveTris[t] = false
			removedTriangles++
		}
	}
	quadrics[e.v1].AddAssign(quadrics[e.v0])
	positions[e.v1] = e.pos
	return removedTriangles
}

// qemEdge is a candidate edge collapse in the priority queue.
type qemEdge struct {
	v0, v1 int
	cost   matrix.Float
	pos    matrix.Vec3
}
