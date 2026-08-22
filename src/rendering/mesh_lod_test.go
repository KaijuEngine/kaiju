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
	lods, err := GenerateMeshLOD(mesh, verts, idx, 3)
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
