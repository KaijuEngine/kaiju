package rendering

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
	"kaijuengine.com/platform/profiler/tracing"
)

const (
	// meshQemChunkTargetVerts is the maximum number of vertices a single chunk may
	// hold before the mesh is split again. It bounds per-goroutine work so chunk
	// processing stays short enough to run at load/runtime without a perceptible
	// hitch.
	meshQemChunkTargetVerts = 512

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

// qemEdgeHeap stores values instead of pointers so rebuilding or refreshing
// candidates does not allocate one object per edge.
type qemEdgeHeap []qemEdge

func (l MeshLod) IsValid() bool { return len(l.Levels) > 0 }

func NewMeshLodGeneratorQem(threads *concurrent.Threads) *MeshLodGeneratorQem {
	return &MeshLodGeneratorQem{
		Threads: threads,
	}
}

// generateMeshLOD preserves the package-local entry point used by focused
// rendering tests and tools that do not have a host thread pool.
func generateMeshLOD(mesh *Mesh, cache *MeshCache, verts []Vertex, indices []uint32, levels int) (MeshLod, error) {
	return NewMeshLodGeneratorQem(nil).GenerateLods(mesh, cache, verts, indices, levels)
}

// processQemChunks preserves the package-local synchronous helper used by
// focused chunking tests.
func processQemChunks(chunks []MeshQemChunk) {
	NewMeshLodGeneratorQem(nil).processQemChunks(chunks)
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
// chunks the mesh once (quadricErrorMetricChunkify), simplifies each chunk
// progressively in parallel, snapshots every requested level, and welds each
// set of chunk snapshots back into a watertight mesh (stitchQemChunks).
//
// The computed LODs are cached on mesh so repeated calls return the same set.
// verts and indices describe the full-resolution geometry; if a level cannot
// be reduced, that level falls back to reusing the source mesh itself.
func (g *MeshLodGeneratorQem) GenerateLods(mesh *Mesh, cache *MeshCache, verts []Vertex, indices []uint32, levels int) (MeshLod, error) {
	defer tracing.NewRegion("MeshLodGeneratorQem.GenerateLods").End()
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
	if levels == 1 {
		mesh.lods = lods
		return lods, nil
	}
	// Chunking, welding, adjacency construction, and the collapse sequence are
	// shared by the entire LOD chain. Re-running QEM from the source for every
	// ratio performs over three times as many collapses for the default five
	// levels and repeatedly allocates the same topology.
	chunks := quadricErrorMetricChunkify(verts, indices, ratios[levels-1])
	if len(chunks) == 0 {
		for i := 1; i < levels; i++ {
			lods.Levels[i] = MeshLODInstance{Mesh: lastLod, Ratio: ratios[i]}
		}
		mesh.lods = lods
		return lods, nil
	}
	levelChunks := g.processQemChunkLevels(chunks, ratios[1:])
	for i := 1; i < levels; i++ {
		ratio := ratios[i]
		lodVerts, lodIndices := stitchQemChunks(verts, levelChunks[i-1])
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
	defer tracing.NewRegion("MeshLodGeneratorQem.processQemChunks").End()
	g.runQemWork(len(chunks), func(i int) {
		chunks[i] = quadricErrorMetricProcessChunk(chunks[i])
	})
}

// processQemChunkLevels initializes each chunk simplifier once and progressively
// advances it through ratios. A worker owns one chunk for the full chain, so no
// simplifier state is shared and no synchronization is needed inside QEM.
func (g *MeshLodGeneratorQem) processQemChunkLevels(chunks []MeshQemChunk, ratios []float32) [][]MeshQemChunk {
	defer tracing.NewRegion("MeshLodGeneratorQem.processQemChunkLevels").End()
	levels := make([][]MeshQemChunk, len(ratios))
	for i := range levels {
		levels[i] = make([]MeshQemChunk, len(chunks))
	}
	g.runQemWork(len(chunks), func(chunkIndex int) {
		simplifier := newQemChunkSimplifier(chunks[chunkIndex])
		triangleCount := len(chunks[chunkIndex].indices) / 3
		for level, ratio := range ratios {
			target := qemTargetTriangles(triangleCount, ratio)
			simplifier.simplifyTo(target)
			levels[level][chunkIndex] = simplifier.snapshot(target)
		}
	})
	return levels
}

func (g *MeshLodGeneratorQem) runQemWork(count int, call func(int)) {
	threads := g.Threads
	if threads == nil || threads.ThreadCount() == 0 || count <= 1 {
		for i := 0; i < count; i++ {
			call(i)
		}
		return
	}
	group := sync.WaitGroup{}
	group.Add(count)
	work := make([]func(int), count)
	for i := range work {
		i := i
		work[i] = func(int) {
			call(i)
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
	defer tracing.NewRegion("stitchQemChunks").End()
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

func qemEdgeLess(a, b qemEdge) bool {
	if a.cost != b.cost {
		return a.cost < b.cost
	}
	if a.v0 != b.v0 {
		return a.v0 < b.v0
	}
	return a.v1 < b.v1
}

func (h *qemEdgeHeap) initialize() {
	for i := len(*h)/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}

func (h *qemEdgeHeap) push(edge qemEdge) {
	*h = append(*h, edge)
	child := len(*h) - 1
	for child > 0 {
		parent := (child - 1) / 2
		if !qemEdgeLess((*h)[child], (*h)[parent]) {
			break
		}
		(*h)[parent], (*h)[child] = (*h)[child], (*h)[parent]
		child = parent
	}
}

func (h *qemEdgeHeap) pop() qemEdge {
	edge := (*h)[0]
	last := len(*h) - 1
	(*h)[0] = (*h)[last]
	(*h)[last] = qemEdge{}
	*h = (*h)[:last]
	if len(*h) > 0 {
		h.down(0)
	}
	return edge
}

func (h *qemEdgeHeap) down(parent int) {
	for {
		left := parent*2 + 1
		if left >= len(*h) {
			return
		}
		child := left
		right := left + 1
		if right < len(*h) && qemEdgeLess((*h)[right], (*h)[left]) {
			child = right
		}
		if !qemEdgeLess((*h)[child], (*h)[parent]) {
			return
		}
		(*h)[parent], (*h)[child] = (*h)[child], (*h)[parent]
		parent = child
	}
}

// buildMeshQemWeldData groups positions that differ only by floating-point seam
// noise. The tolerance is relative to the mesh diagonal and deliberately small:
// it joins duplicated attribute vertices without merging visibly distinct
// geometry.
func buildMeshQemWeldData(verts []Vertex) meshQemWeldData {
	defer tracing.NewRegion("buildMeshQemWeldData").End()
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
// Chunks grow breadth-first through welded triangle adjacency until their local
// vertex count reaches meshQemChunkTargetVerts. This minimizes cut vertices
// compared with slicing arbitrary index order. A vertex that spans more than
// one chunk is duplicated and flagged as boundary, keeping shared seams
// watertight while chunks simplify independently.
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
	defer tracing.NewRegion("quadricErrorMetricChunkify").End()
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
	tris := len(indices) / 3
	if len(welded.positions) <= meshQemChunkTargetVerts {
		chunk := MeshQemChunk{}
		localToGlobal := make([]int, 0, len(welded.positions))
		globalToLocal := make(map[int]int, len(welded.positions))
		for corner := 0; corner < tris*3; corner++ {
			source := int(indices[corner])
			if source < 0 || source >= len(welded.sourceToWeld) {
				return nil
			}
			geometry := welded.sourceToWeld[source]
			local, exists := globalToLocal[geometry]
			if !exists {
				local = len(localToGlobal)
				localToGlobal = append(localToGlobal, geometry)
				globalToLocal[geometry] = local
			}
			chunk.indices = append(chunk.indices, uint32(local))
			chunk.sourceIndices = append(chunk.sourceIndices, source)
		}
		emitQemChunk(&chunks, &chunk, localToGlobal, vertexSlots, ratio)
		buildQemChunkGeometry(welded, chunks, vertexSlots)
		return chunks
	}
	triangleGeometry := make([][3]int, tris)
	vertexTriangleCounts := make([]int, len(welded.positions))
	for t := 0; t < tris; t++ {
		base := t * 3
		for corner := 0; corner < 3; corner++ {
			source := int(indices[base+corner])
			if source < 0 || source >= len(welded.sourceToWeld) {
				return nil
			}
			geometry := welded.sourceToWeld[source]
			triangleGeometry[t][corner] = geometry
			unique := true
			for previous := 0; previous < corner; previous++ {
				if triangleGeometry[t][previous] == geometry {
					unique = false
					break
				}
			}
			if unique {
				vertexTriangleCounts[geometry]++
			}
		}
	}
	// A compact vertex-to-triangle CSR table lets breadth-first partitioning
	// traverse topology without thousands of per-vertex map/slice allocations.
	vertexTriangleOffsets := make([]int, len(welded.positions)+1)
	for vertex, count := range vertexTriangleCounts {
		vertexTriangleOffsets[vertex+1] = vertexTriangleOffsets[vertex] + count
	}
	vertexTriangles := make([]int, vertexTriangleOffsets[len(vertexTriangleOffsets)-1])
	vertexTriangleCursor := append([]int(nil), vertexTriangleOffsets[:len(welded.positions)]...)
	for triangle, geometry := range triangleGeometry {
		for corner, vertex := range geometry {
			unique := true
			for previous := 0; previous < corner; previous++ {
				if geometry[previous] == vertex {
					unique = false
					break
				}
			}
			if unique {
				vertexTriangles[vertexTriangleCursor[vertex]] = triangle
				vertexTriangleCursor[vertex]++
			}
		}
	}
	visited := make([]bool, tris)
	queued := make([]uint32, tris)
	queue := make([]int, 0, min(tris, meshQemChunkTargetVerts*2))
	queueGeneration := uint32(0)
	localToGlobal := make([]int, 0, meshQemChunkTargetVerts)
	globalToLocal := make(map[int]int, meshQemChunkTargetVerts)
	// cornerLocal holds the remapped local index of each corner of the triangle
	// currently being processed.
	cornerLocal := [3]int{}
	current := MeshQemChunk{}
	addTriangle := func(triangle int) {
		base := triangle * 3
		for c := 0; c < 3; c++ {
			source := int(indices[base+c])
			g := triangleGeometry[triangle][c]
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
	for seed := 0; seed < tris; seed++ {
		if visited[seed] {
			continue
		}
		queueGeneration++
		if queueGeneration == 0 {
			clear(queued)
			queueGeneration = 1
		}
		queue = queue[:0]
		queue = append(queue, seed)
		queued[seed] = queueGeneration
		for head := 0; head < len(queue); head++ {
			triangle := queue[head]
			if visited[triangle] {
				continue
			}
			geometry := triangleGeometry[triangle]
			newVertices := 0
			for corner, vertex := range geometry {
				unique := true
				for previous := 0; previous < corner; previous++ {
					if geometry[previous] == vertex {
						unique = false
						break
					}
				}
				if unique {
					if _, exists := globalToLocal[vertex]; !exists {
						newVertices++
					}
				}
			}
			if len(current.indices) > 0 && len(localToGlobal)+newVertices > meshQemChunkTargetVerts {
				continue
			}
			visited[triangle] = true
			addTriangle(triangle)
			for corner, vertex := range geometry {
				unique := true
				for previous := 0; previous < corner; previous++ {
					if geometry[previous] == vertex {
						unique = false
						break
					}
				}
				if !unique {
					continue
				}
				for _, adjacent := range vertexTriangles[vertexTriangleOffsets[vertex]:vertexTriangleOffsets[vertex+1]] {
					if !visited[adjacent] && queued[adjacent] != queueGeneration {
						queued[adjacent] = queueGeneration
						queue = append(queue, adjacent)
					}
				}
			}
		}
		emitQemChunk(&chunks, &current, localToGlobal, vertexSlots, ratio)
		current = MeshQemChunk{}
		clear(globalToLocal)
		localToGlobal = localToGlobal[:0]
	}
	buildQemChunkGeometry(welded, chunks, vertexSlots)
	return chunks
}

// emitQemChunk finalizes a collected chunk: it records each member vertex in
// vertexSlots (so boundary membership can be resolved later) and computes the
// chunk's targetTriangles from its own triangle count and the requested ratio.
func emitQemChunk(chunks *[]MeshQemChunk, chunk *MeshQemChunk, localToGlobal []int, vertexSlots [][]meshQemSlot, ratio float32) {
	defer tracing.NewRegion("emitQemChunk").End()
	chunkIndex := len(*chunks)
	for li, g := range localToGlobal {
		vertexSlots[g] = append(vertexSlots[g], meshQemSlot{chunk: chunkIndex, local: li})
	}
	chunk.targetTriangles = qemTargetTriangles(len(chunk.indices)/3, ratio)
	*chunks = append(*chunks, *chunk)
}

func qemTargetTriangles(triangleCount int, ratio float32) int {
	if ratio >= 1 {
		return triangleCount
	}
	target := int(float32(triangleCount) * ratio)
	if triangleCount > 0 && target < 1 {
		return 1
	}
	if target > triangleCount {
		return triangleCount
	}
	return target
}

// buildQemChunkGeometry fills in each chunk's local welded positions, source
// variants, and boundary flags. A vertex is on the boundary iff it appears in
// more than one chunk.
func buildQemChunkGeometry(welded meshQemWeldData, chunks []MeshQemChunk, vertexSlots [][]meshQemSlot) {
	defer tracing.NewRegion("buildQemChunkGeometry").End()
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
	defer tracing.NewRegion("selectMeshLodRatios").End()
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
	defer tracing.NewRegion("quadricErrorMetricProcessChunk").End()
	if len(chunk.indices) < 3 || len(chunk.positions) == 0 {
		chunk.resultIndices = append(chunk.resultIndices[:0], chunk.indices...)
		chunk.resultSources = append(chunk.resultSources[:0], chunk.sourceIndices...)
		return chunk
	}
	simplifier := newQemChunkSimplifier(chunk)
	simplifier.simplifyTo(chunk.targetTriangles)
	return simplifier.snapshot(chunk.targetTriangles)
}

// qemChunkSimplifier owns the mutable state for one chunk. Adjacency and the
// edge queue are built once. Each collapse rewires only incident triangles and
// refreshes candidates in the affected one-ring; old heap entries are rejected
// cheaply through per-vertex versions.
type qemChunkSimplifier struct {
	source           MeshQemChunk
	positions        []matrix.Vec3
	normals          []matrix.Vec3
	triangles        [][3]int
	triangleSources  [][3]int
	liveTris         []bool
	liveVertices     []bool
	referenceFaces   []matrix.Vec3
	quadrics         []matrix.Mat4
	incident         [][]int
	protected        []bool
	topologyBoundary []bool
	versions         []uint32
	liveCount        int
	areaEpsilon      matrix.Float
	edges            qemEdgeHeap
	lastQueued       map[uint64]uint32
	queueGeneration  uint32
	vertexMarks      []uint32
	commonMarks      []uint32
	triangleMarks    []uint32
	vertexMark       uint32
	commonMark       uint32
	triangleMark     uint32
	affected         []int
	stuck            bool
}

func newQemChunkSimplifier(chunk MeshQemChunk) *qemChunkSimplifier {
	vertexCount := len(chunk.positions)
	triangleCount := len(chunk.indices) / 3
	s := &qemChunkSimplifier{
		source:           chunk,
		positions:        append([]matrix.Vec3(nil), chunk.positions...),
		normals:          make([]matrix.Vec3, vertexCount),
		triangles:        make([][3]int, triangleCount),
		triangleSources:  make([][3]int, triangleCount),
		liveTris:         make([]bool, triangleCount),
		liveVertices:     make([]bool, vertexCount),
		referenceFaces:   make([]matrix.Vec3, triangleCount),
		quadrics:         make([]matrix.Mat4, vertexCount),
		incident:         make([][]int, vertexCount),
		protected:        make([]bool, vertexCount),
		topologyBoundary: make([]bool, vertexCount),
		versions:         make([]uint32, vertexCount),
		lastQueued:       make(map[uint64]uint32, triangleCount*2),
		vertexMarks:      make([]uint32, vertexCount),
		commonMarks:      make([]uint32, vertexCount),
		triangleMarks:    make([]uint32, triangleCount),
		affected:         make([]int, 0, 32),
	}
	copy(s.normals, chunk.normals)
	copy(s.protected, chunk.boundary)
	for i := range s.liveVertices {
		s.liveVertices[i] = true
		s.versions[i] = 1
	}
	s.areaEpsilon = qemAreaEpsilon(s.positions)
	edgeUses := make(map[uint64]uint8, triangleCount*2)
	incidentCounts := make([]int, vertexCount)
	for t := 0; t < triangleCount; t++ {
		a := int(chunk.indices[t*3])
		b := int(chunk.indices[t*3+1])
		c := int(chunk.indices[t*3+2])
		tri := [3]int{a, b, c}
		s.triangles[t] = tri
		for corner, local := range tri {
			source := -1
			inputCorner := t*3 + corner
			if inputCorner < len(chunk.sourceIndices) {
				source = chunk.sourceIndices[inputCorner]
			} else if local >= 0 && local < len(chunk.sourceVariants) && len(chunk.sourceVariants[local]) > 0 {
				source = chunk.sourceVariants[local][0]
			}
			s.triangleSources[t][corner] = source
		}
		if a < 0 || b < 0 || c < 0 || a >= vertexCount || b >= vertexCount || c >= vertexCount ||
			a == b || b == c || a == c {
			continue
		}
		face := s.positions[b].Subtract(s.positions[a]).Cross(s.positions[c].Subtract(s.positions[a]))
		if face.LengthSquared() <= s.areaEpsilon {
			continue
		}
		s.referenceFaces[t] = face
		s.liveTris[t] = true
		s.liveCount++
		incidentCounts[a]++
		incidentCounts[b]++
		incidentCounts[c]++
		addPlaneQuadric(&s.quadrics[a], s.positions[a], s.positions[b], s.positions[c])
		addPlaneQuadric(&s.quadrics[b], s.positions[a], s.positions[b], s.positions[c])
		addPlaneQuadric(&s.quadrics[c], s.positions[a], s.positions[b], s.positions[c])
		edgeUses[qemEdgeKey(a, b)]++
		edgeUses[qemEdgeKey(b, c)]++
		edgeUses[qemEdgeKey(c, a)]++
	}
	for vertex, count := range incidentCounts {
		if count > 0 {
			// A little spare capacity absorbs the usual one-ring transfers without
			// forcing a new allocation on every successful collapse.
			s.incident[vertex] = make([]int, 0, count+4)
		}
	}
	for t, tri := range s.triangles {
		if !s.liveTris[t] {
			continue
		}
		s.incident[tri[0]] = append(s.incident[tri[0]], t)
		s.incident[tri[1]] = append(s.incident[tri[1]], t)
		s.incident[tri[2]] = append(s.incident[tri[2]], t)
	}
	// Inter-chunk vertices remain fixed. Original topological-boundary vertices
	// can collapse only along the same kind of boundary, matching the previous
	// simplifier's restrictions.
	for key, uses := range edgeUses {
		if uses != 2 {
			a, b := qemEdgeVertices(key)
			s.topologyBoundary[a] = true
			s.topologyBoundary[b] = true
		}
	}
	s.rebuildEdgeQueue()
	return s
}

func (s *qemChunkSimplifier) simplifyTo(target int) {
	if target < 0 {
		target = 0
	}
	for s.liveCount > target && !s.stuck {
		collapsed := false
		for len(s.edges) > 0 {
			edge := s.edges.pop()
			if !s.edgeIsCurrent(edge) || !s.collapsePreservesTopology(edge) || !s.collapsePreservesGeometry(edge) {
				continue
			}
			affected := s.collectAffected(edge.v0, edge.v1)
			removed := s.applyCollapse(edge)
			if removed == 0 {
				continue
			}
			s.liveCount -= removed
			s.refreshAffected(affected)
			collapsed = true
			break
		}
		if !collapsed {
			s.stuck = true
		}
	}
}

func (s *qemChunkSimplifier) snapshot(target int) MeshQemChunk {
	chunk := s.source
	chunk.targetTriangles = target
	chunk.positions = append([]matrix.Vec3(nil), s.positions...)
	chunk.quadrics = append([]matrix.Mat4(nil), s.quadrics...)
	chunk.resultIndices = make([]uint32, 0, s.liveCount*3)
	chunk.resultSources = make([]int, 0, s.liveCount*3)
	for t, tri := range s.triangles {
		if !s.liveTris[t] {
			continue
		}
		chunk.resultIndices = append(chunk.resultIndices, uint32(tri[0]), uint32(tri[1]), uint32(tri[2]))
		chunk.resultSources = append(chunk.resultSources,
			s.triangleSources[t][0], s.triangleSources[t][1], s.triangleSources[t][2])
	}
	return chunk
}

func (s *qemChunkSimplifier) rebuildEdgeQueue() {
	s.edges = s.edges[:0]
	s.nextQueueGeneration()
	for t, tri := range s.triangles {
		if !s.liveTris[t] {
			continue
		}
		s.queueEdge(tri[0], tri[1], false)
		s.queueEdge(tri[1], tri[2], false)
		s.queueEdge(tri[2], tri[0], false)
	}
	s.edges.initialize()
}

func (s *qemChunkSimplifier) queueEdge(a, b int, initialized bool) {
	if a == b || !s.liveVertices[a] || !s.liveVertices[b] ||
		s.topologyBoundary[a] != s.topologyBoundary[b] || s.protected[a] && s.protected[b] {
		return
	}
	key := qemEdgeKey(a, b)
	if s.lastQueued[key] == s.queueGeneration {
		return
	}
	s.lastQueued[key] = s.queueGeneration
	v0, v1 := qemEdgeVertices(key)
	if s.protected[v0] {
		v0, v1 = v1, v0
	}
	combined := s.quadrics[v0]
	combined.AddAssign(s.quadrics[v1])
	cost, position := qemCollapsePosition(combined, s.positions[v0], s.positions[v1], s.protected[v1])
	if matrix.IsNaN(cost) || matrix.IsInf(cost, 0) || position.IsNaN() || position.IsInf(0) {
		return
	}
	edge := qemEdge{
		v0:       v0,
		v1:       v1,
		cost:     cost,
		pos:      position,
		version0: s.versions[v0],
		version1: s.versions[v1],
	}
	if initialized {
		s.edges.push(edge)
	} else {
		s.edges = append(s.edges, edge)
	}
}

func (s *qemChunkSimplifier) edgeIsCurrent(edge qemEdge) bool {
	return s.liveVertices[edge.v0] && s.liveVertices[edge.v1] &&
		s.versions[edge.v0] == edge.version0 && s.versions[edge.v1] == edge.version1
}

func (s *qemChunkSimplifier) collapsePreservesTopology(edge qemEdge) bool {
	neighborMark := s.nextVertexMark()
	edgeUses := 0
	for _, t := range s.incident[edge.v0] {
		if !s.liveTris[t] {
			continue
		}
		tri := s.triangles[t]
		usesSurvivor := false
		for _, vertex := range tri {
			if vertex == edge.v1 {
				usesSurvivor = true
			}
			if vertex != edge.v0 {
				s.vertexMarks[vertex] = neighborMark
			}
		}
		if usesSurvivor {
			edgeUses++
		}
	}
	if edgeUses < 1 || edgeUses > 2 {
		return false
	}
	countedMark := s.nextCommonMark()
	common := 0
	for _, t := range s.incident[edge.v1] {
		if !s.liveTris[t] {
			continue
		}
		for _, vertex := range s.triangles[t] {
			if vertex != edge.v1 && s.vertexMarks[vertex] == neighborMark && s.commonMarks[vertex] != countedMark {
				s.commonMarks[vertex] = countedMark
				common++
			}
		}
	}
	return common == edgeUses
}

func (s *qemChunkSimplifier) collapsePreservesGeometry(edge qemEdge) bool {
	triangleMark := s.nextTriangleMark()
	endpoints := [2]int{edge.v0, edge.v1}
	for _, endpoint := range endpoints {
		for _, t := range s.incident[endpoint] {
			if !s.liveTris[t] || s.triangleMarks[t] == triangleMark {
				continue
			}
			s.triangleMarks[t] = triangleMark
			tri := s.triangles[t]
			usesRemoved := qemTriangleContains(tri, edge.v0)
			usesSurvivor := qemTriangleContains(tri, edge.v1)
			if usesRemoved && usesSurvivor {
				continue
			}
			updated := [3]matrix.Vec3{s.positions[tri[0]], s.positions[tri[1]], s.positions[tri[2]]}
			updatedIndices := tri
			for i, vertex := range tri {
				if vertex == edge.v0 || vertex == edge.v1 {
					updated[i] = edge.pos
				}
				if vertex == edge.v0 {
					updatedIndices[i] = edge.v1
				}
			}
			newFace := updated[1].Subtract(updated[0]).Cross(updated[2].Subtract(updated[0]))
			newArea := newFace.LengthSquared()
			referenceArea := s.referenceFaces[t].LengthSquared()
			direction := s.referenceFaces[t].Dot(newFace)
			const minNormalDotSquared = matrix.Float(0.04)
			if newFace.IsNaN() || newFace.IsInf(0) || newArea <= s.areaEpsilon || direction <= 0 ||
				direction*direction < referenceArea*newArea*minNormalDotSquared {
				return false
			}
			vertexNormal := s.normals[updatedIndices[0]].Add(s.normals[updatedIndices[1]]).Add(s.normals[updatedIndices[2]])
			if !vertexNormal.IsZero() {
				normalDirection := vertexNormal.Dot(newFace)
				if normalDirection <= 0 || normalDirection*normalDirection < vertexNormal.LengthSquared()*newArea*minNormalDotSquared {
					return false
				}
			}
		}
	}
	return true
}

func (s *qemChunkSimplifier) collectAffected(removed, survivor int) []int {
	mark := s.nextVertexMark()
	s.affected = s.affected[:0]
	add := func(vertex int) {
		if s.vertexMarks[vertex] == mark {
			return
		}
		s.vertexMarks[vertex] = mark
		s.affected = append(s.affected, vertex)
	}
	add(removed)
	add(survivor)
	for _, endpoint := range [2]int{removed, survivor} {
		for _, t := range s.incident[endpoint] {
			if !s.liveTris[t] {
				continue
			}
			for _, vertex := range s.triangles[t] {
				add(vertex)
			}
		}
	}
	return s.affected
}

func (s *qemChunkSimplifier) applyCollapse(edge qemEdge) int {
	removedTriangles := 0
	for _, t := range s.incident[edge.v0] {
		if !s.liveTris[t] {
			continue
		}
		tri := s.triangles[t]
		if qemTriangleContains(tri, edge.v1) {
			s.liveTris[t] = false
			removedTriangles++
			continue
		}
		for i := range tri {
			if tri[i] == edge.v0 {
				tri[i] = edge.v1
			}
		}
		s.triangles[t] = tri
		s.incident[edge.v1] = append(s.incident[edge.v1], t)
	}
	if removedTriangles == 0 {
		return 0
	}
	s.quadrics[edge.v1].AddAssign(s.quadrics[edge.v0])
	s.positions[edge.v1] = edge.pos
	s.liveVertices[edge.v0] = false
	s.incident[edge.v0] = nil
	return removedTriangles
}

func (s *qemChunkSimplifier) refreshAffected(affected []int) {
	for _, vertex := range affected {
		s.versions[vertex]++
		if !s.liveVertices[vertex] {
			continue
		}
		incident := s.incident[vertex][:0]
		for _, t := range s.incident[vertex] {
			if s.liveTris[t] && qemTriangleContains(s.triangles[t], vertex) {
				incident = append(incident, t)
			}
		}
		s.incident[vertex] = incident
	}
	// Lazy invalidation normally keeps refreshes local. Rebuild occasionally if
	// stale candidates grow far beyond the current topology.
	if len(s.edges) > max(128, s.liveCount*6) {
		s.rebuildEdgeQueue()
		return
	}
	s.nextQueueGeneration()
	for _, vertex := range affected {
		if !s.liveVertices[vertex] {
			continue
		}
		for _, t := range s.incident[vertex] {
			tri := s.triangles[t]
			for _, neighbor := range tri {
				if neighbor != vertex {
					s.queueEdge(vertex, neighbor, true)
				}
			}
		}
	}
}

func (s *qemChunkSimplifier) nextQueueGeneration() {
	s.queueGeneration++
	if s.queueGeneration == 0 {
		clear(s.lastQueued)
		s.queueGeneration = 1
	}
}

func (s *qemChunkSimplifier) nextVertexMark() uint32 {
	s.vertexMark++
	if s.vertexMark == 0 {
		clear(s.vertexMarks)
		s.vertexMark = 1
	}
	return s.vertexMark
}

func (s *qemChunkSimplifier) nextCommonMark() uint32 {
	s.commonMark++
	if s.commonMark == 0 {
		clear(s.commonMarks)
		s.commonMark = 1
	}
	return s.commonMark
}

func (s *qemChunkSimplifier) nextTriangleMark() uint32 {
	s.triangleMark++
	if s.triangleMark == 0 {
		clear(s.triangleMarks)
		s.triangleMark = 1
	}
	return s.triangleMark
}

func qemEdgeKey(a, b int) uint64 {
	if a > b {
		a, b = b, a
	}
	return uint64(uint32(a))<<32 | uint64(uint32(b))
}

func qemEdgeVertices(key uint64) (int, int) {
	return int(uint32(key >> 32)), int(uint32(key))
}

func qemTriangleContains(triangle [3]int, vertex int) bool {
	return triangle[0] == vertex || triangle[1] == vertex || triangle[2] == vertex
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

// qemEdge is a candidate edge collapse in the priority queue.
type qemEdge struct {
	v0, v1             int
	cost               matrix.Float
	pos                matrix.Vec3
	version0, version1 uint32
}
