package rendering

import (
	"testing"

	"kaijuengine.com/matrix"
)

// buildGrid2 builds an n x n grid of quads (two triangles each) laid out in the
// XY plane, giving a flat, easily-simplifiable mesh for testing LOD generation.
func buildGrid2(n int) ([]Vertex, []uint32) {
	verts := make([]Vertex, 0, (n+1)*(n+1))
	idx := make([]uint32, 0, n*n*6)
	for y := 0; y <= n; y++ {
		for x := 0; x <= n; x++ {
			verts = append(verts, Vertex{Position: matrix.Vec3{float32(x), float32(y), 0}})
		}
	}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			a := uint32(y*(n+1) + x)
			b := uint32(y*(n+1) + x + 1)
			c := uint32((y+1)*(n+1) + x)
			d := uint32((y+1)*(n+1) + x + 1)
			idx = append(idx, a, b, c, b, d, c)
		}
	}
	return verts, idx
}

func TestGenerateMeshLODReducesTriangles(t *testing.T) {
	verts, idx := buildGrid2(8)
	mesh := NewMesh("test-grid", verts, idx)
	cache := NewMeshCache(nil, nil)
	lods, err := generateMeshLOD(mesh, &cache, verts, idx, 3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !lods.IsValid() || len(lods.Levels) != 3 {
		t.Fatalf("expected 3 valid levels, got %#v", lods)
	}
	if lods.Levels[0].Mesh != mesh {
		t.Fatalf("level 0 should be the source mesh")
	}
	full := len(idx) / 3
	prevTri := full
	for i := 0; i < len(lods.Levels); i++ {
		m := lods.Levels[i].Mesh
		triCount := full
		if m != mesh {
			triCount = len(m.pendingIndexes) / 3
			if triCount > prevTri {
				t.Errorf("level %d (%d) has more triangles than previous (%d)", i, triCount, prevTri)
			}
		}
		prevTri = triCount
	}
	// Ensure later ratios genuinely reduced the mesh (not just reused it).
	if lods.Levels[len(lods.Levels)-1].Mesh == mesh {
		t.Errorf("last level unexpectedly reused the full mesh")
	}
}

// buildCube builds an 8-vertex, 12-triangle cube (the same layout used by
// NewMeshCube). It is used to test the already-low-poly short-circuit.
func buildCube() ([]Vertex, []uint32) {
	verts := make([]Vertex, 8)
	verts[0].Position = matrix.Vec3{-0.5, -0.5, 0.5}
	verts[1].Position = matrix.Vec3{-0.5, 0.5, 0.5}
	verts[2].Position = matrix.Vec3{0.5, 0.5, 0.5}
	verts[3].Position = matrix.Vec3{0.5, -0.5, 0.5}
	verts[4].Position = matrix.Vec3{-0.5, -0.5, -0.5}
	verts[5].Position = matrix.Vec3{-0.5, 0.5, -0.5}
	verts[6].Position = matrix.Vec3{0.5, 0.5, -0.5}
	verts[7].Position = matrix.Vec3{0.5, -0.5, -0.5}
	indexes := []uint32{
		5, 2, 6, 2, 0, 3,
		1, 4, 0, 7, 0, 4,
		6, 3, 7, 5, 7, 4,
		5, 1, 2, 2, 1, 0,
		1, 5, 4, 7, 3, 0,
		6, 2, 3, 5, 6, 7,
	}
	return verts, indexes
}

// TestGenerateMeshLODSkipsLowPolyMesh verifies that a mesh already at or below
// the low-poly floor reuses the source mesh for every LOD level, without
// generating any additional simplified meshes.
func TestGenerateMeshLODSkipsLowPolyMesh(t *testing.T) {
	verts, idx := buildCube() // 12 triangles, well under the floor
	mesh := NewMesh("test-cube", verts, idx)
	cache := NewMeshCache(nil, nil)
	lods, err := generateMeshLOD(mesh, &cache, verts, idx, 5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !lods.IsValid() || len(lods.Levels) != 5 {
		t.Fatalf("expected 5 valid levels, got %#v", lods)
	}
	for i, lvl := range lods.Levels {
		if lvl.Mesh != mesh {
			t.Errorf("level %d should reuse the source mesh, got %v", i, lvl.Mesh)
		}
	}
	// No LOD meshes should have been created in the cache.
	if len(cache.meshes) != 0 {
		t.Errorf("expected no LOD meshes cached, got %d", len(cache.meshes))
	}
}

// TestGenerateMeshLODClampsRemainingLevels verifies that once a generated LOD
// reaches the low-poly floor, every remaining level reuses that same LOD mesh
// instead of continuing to simplify.
func TestGenerateMeshLODClampsRemainingLevels(t *testing.T) {
	verts, idx := buildGrid2(8) // 128 triangles, well above the floor
	mesh := NewMesh("test-grid", verts, idx)
	cache := NewMeshCache(nil, nil)
	lods, err := generateMeshLOD(mesh, &cache, verts, idx, 6)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !lods.IsValid() || len(lods.Levels) != 6 {
		t.Fatalf("expected 6 valid levels, got %#v", lods)
	}
	// Find the first level that reached the floor.
	floorIdx := -1
	for i := 1; i < len(lods.Levels); i++ {
		m := lods.Levels[i].Mesh
		if m == mesh {
			continue
		}
		if len(m.pendingIndexes)/3 <= meshLodMinTriangles {
			floorIdx = i
			break
		}
	}
	if floorIdx == -1 {
		t.Fatalf("expected some level to reach the low-poly floor (<=%d tris)", meshLodMinTriangles)
	}
	// All levels after the floor must reuse the exact same last LOD mesh.
	lastLod := lods.Levels[floorIdx].Mesh
	for j := floorIdx + 1; j < len(lods.Levels); j++ {
		if lods.Levels[j].Mesh != lastLod {
			t.Errorf("level %d should reuse the floor LOD mesh, got a different mesh", j)
		}
	}
}

func TestGenerateMeshLODSphereHasNoDegenerateTriangles(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	sphere := NewMeshSphere(&cache, 1, 32, 32)
	if !sphere.lods.IsValid() {
		t.Fatal("expected sphere LODs to be generated")
	}
	for level := 1; level < len(sphere.lods.Levels); level++ {
		lod := sphere.lods.Levels[level].Mesh
		t.Logf("LOD %d: %d vertices, %d triangles", level, len(lod.pendingVerts), len(lod.pendingIndexes)/3)
		for i := 0; i < len(lod.pendingIndexes); i += 3 {
			a := lod.pendingIndexes[i]
			b := lod.pendingIndexes[i+1]
			c := lod.pendingIndexes[i+2]
			if a == b || b == c || a == c {
				t.Fatalf("LOD %d triangle %d has duplicate indices: %d, %d, %d", level, i/3, a, b, c)
			}
			p0 := lod.pendingVerts[a].Position
			p1 := lod.pendingVerts[b].Position
			p2 := lod.pendingVerts[c].Position
			face := p1.Subtract(p0).Cross(p2.Subtract(p0))
			if face.LengthSquared() <= 1e-12 {
				t.Fatalf("LOD %d triangle %d has zero area", level, i/3)
			}
			centroid := p0.Add(p1).Add(p2).Scale(1.0 / 3.0)
			if face.Dot(centroid) <= 0 {
				t.Fatalf("LOD %d triangle %d faces inward: indices=%v positions=%v,%v,%v face=%v centroid=%v", level, i/3,
					[]uint32{a, b, c}, p0, p1, p2, face, centroid)
			}
		}
	}
}

func TestQuadricErrorMetricKeepsProtectedVerticesFixed(t *testing.T) {
	chunk := MeshQemChunk{
		positions: []matrix.Vec3{
			{0, 0, 0},
			{-1, 0, -1},
			{1, 0, -1},
			{1, 0, 1},
			{-1, 0, 1},
		},
		indices: []uint32{
			0, 2, 1,
			0, 3, 2,
			0, 4, 3,
			0, 1, 4,
		},
		boundary:        []bool{false, true, true, true, true},
		globalIndices:   []int{0, 1, 2, 3, 4},
		targetTriangles: 2,
	}
	want := append([]matrix.Vec3(nil), chunk.positions...)
	result := quadricErrorMetricProcessChunk(chunk)
	for i := 1; i < len(want); i++ {
		if result.positions[i] != want[i] {
			t.Errorf("protected vertex %d moved from %v to %v", i, want[i], result.positions[i])
		}
	}
}

func TestQuadricErrorMetricChunkifyDoesNotDuplicateSplitTriangle(t *testing.T) {
	vertexCount := meshQemChunkTargetVerts + 2
	verts := make([]Vertex, vertexCount)
	indices := make([]uint32, 0, (vertexCount/3)*3)
	for i := 0; i+2 < vertexCount; i += 3 {
		indices = append(indices, uint32(i), uint32(i+1), uint32(i+2))
	}
	chunks := quadricErrorMetricChunkify(verts, indices, 0.5)
	gotTriangles := 0
	for i := range chunks {
		gotTriangles += len(chunks[i].indices) / 3
	}
	if wantTriangles := len(indices) / 3; gotTriangles != wantTriangles {
		t.Fatalf("chunking produced %d triangles, want %d", gotTriangles, wantTriangles)
	}
}
