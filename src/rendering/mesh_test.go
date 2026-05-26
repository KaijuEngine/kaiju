/******************************************************************************/
/* mesh_test.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"

	"kaijuengine.com/matrix"
)

func assertMeshCounts(t *testing.T, mesh *Mesh, verts, indexes int) {
	t.Helper()
	if len(mesh.pendingVerts) != verts {
		t.Fatalf("%s verts = %d, want %d", mesh.Key(), len(mesh.pendingVerts), verts)
	}
	if len(mesh.pendingIndexes) != indexes {
		t.Fatalf("%s indexes = %d, want %d", mesh.Key(), len(mesh.pendingIndexes), indexes)
	}
}

func assertMeshFacesFollowVertexNormals(t *testing.T, mesh *Mesh) {
	t.Helper()
	checked := 0
	for i := 0; i < len(mesh.pendingIndexes); i += 3 {
		tri := [3]Vertex{
			mesh.pendingVerts[mesh.pendingIndexes[i]],
			mesh.pendingVerts[mesh.pendingIndexes[i+1]],
			mesh.pendingVerts[mesh.pendingIndexes[i+2]],
		}
		faceNormal := VertexFaceNormal(tri)
		if faceNormal.IsZero() {
			continue
		}
		vertexNormal := tri[0].Normal.Add(tri[1].Normal).Add(tri[2].Normal)
		if vertexNormal.IsZero() {
			continue
		}
		if faceNormal.Dot(vertexNormal.Normal()) <= 0 {
			t.Fatalf("%s triangle %d winding opposes vertex normals: face=%v vertex=%v indexes=%v",
				mesh.Key(), i/3, faceNormal, vertexNormal.Normal(), mesh.pendingIndexes[i:i+3])
		}
		checked++
	}
	if checked == 0 {
		t.Fatalf("%s had no non-degenerate triangle to test", mesh.Key())
	}
}

func TestNewMeshComputesBounds(t *testing.T) {
	verts := []Vertex{
		{Position: matrix.Vec3{-1, 2, -3}},
		{Position: matrix.Vec3{5, -6, 7}},
		{Position: matrix.Vec3{2, 3, 4}},
	}
	mesh := NewMesh("bounds", verts, []uint32{0, 1, 2})
	if mesh.Key() != "bounds" {
		t.Fatalf("Key = %q", mesh.Key())
	}
	if len(mesh.pendingVerts) != len(verts) || len(mesh.pendingIndexes) != 3 {
		t.Fatalf("pending mesh data not retained")
	}
	bounds := mesh.Bounds()
	if got := bounds.Min(); got != (matrix.Vec3{-1, -6, -3}) {
		t.Fatalf("bounds min = %v", got)
	}
	if got := bounds.Max(); got != (matrix.Vec3{5, 3, 7}) {
		t.Fatalf("bounds max = %v", got)
	}
	empty := NewMesh("empty", nil, nil)
	if empty.Bounds().Center != matrix.Vec3Zero() || empty.Bounds().Extent != matrix.Vec3Zero() {
		t.Fatalf("empty mesh bounds = %+v", empty.Bounds())
	}
}

func TestNewDynamicMeshMarksDynamic(t *testing.T) {
	mesh := NewDynamicMesh("dynamic", testVerts(), []uint32{0, 1})
	if !mesh.dynamic {
		t.Fatalf("dynamic flag was not set")
	}
	assertMeshCounts(t, mesh, 2, 2)
}

func TestMeshQuadData(t *testing.T) {
	verts, indexes := MeshQuadData()
	if len(verts) != 4 || len(indexes) != 6 {
		t.Fatalf("quad counts = %d/%d", len(verts), len(indexes))
	}
	if verts[0].Position != (matrix.Vec3{-0.5, -0.5, 0}) ||
		verts[2].Position != (matrix.Vec3{0.5, 0.5, 0}) {
		t.Fatalf("unexpected quad positions: %+v", verts)
	}
	for i := range verts {
		if verts[i].Normal != (matrix.Vec3{0, 0, 1}) || verts[i].Color != matrix.ColorWhite() {
			t.Fatalf("unexpected quad vertex %d: %+v", i, verts[i])
		}
	}
	want := []uint32{0, 2, 1, 0, 3, 2}
	for i := range want {
		if indexes[i] != want[i] {
			t.Fatalf("quad index[%d] = %d, want %d", i, indexes[i], want[i])
		}
	}
}

func TestMeshPlaneData(t *testing.T) {
	verts, indexes := MeshPlaneData()
	if len(verts) != 4 || len(indexes) != 6 {
		t.Fatalf("plane counts = %d/%d", len(verts), len(indexes))
	}
	for i := range verts {
		if verts[i].Normal != matrix.Vec3Up() || verts[i].Color != matrix.ColorWhite() {
			t.Fatalf("unexpected plane vertex %d: %+v", i, verts[i])
		}
	}
	if verts[0].Position != (matrix.Vec3{-0.5, 0, 0.5}) ||
		verts[2].Position != (matrix.Vec3{0.5, 0, -0.5}) {
		t.Fatalf("unexpected plane positions: %+v", verts)
	}
}

func TestMeshPrimitiveGeneratorsReuseCache(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	cases := []struct {
		name    string
		create  func() *Mesh
		verts   int
		indexes int
	}{
		{"quad", func() *Mesh { return NewMeshQuad(&cache) }, 4, 6},
		{"triangle", func() *Mesh { return NewMeshTriangle(&cache) }, 3, 3},
		{"unit_quad", func() *Mesh { return NewMeshUnitQuad(&cache) }, 4, 6},
		{"screen_quad", func() *Mesh { return NewMeshScreenQuad(&cache) }, 4, 6},
		{"plane", func() *Mesh { return NewMeshPlane(&cache) }, 4, 6},
		{"cube", func() *Mesh { return NewMeshCube(&cache) }, 8, 36},
		{"skybox", func() *Mesh { return NewMeshSkyboxCube(&cache) }, 8, 36},
		{"cube_inverse", func() *Mesh { return NewMeshCubeInverse(&cache) }, 8, 36},
	}
	for _, c := range cases {
		first := c.create()
		second := c.create()
		if first != second {
			t.Fatalf("%s did not reuse cached mesh", c.name)
		}
		assertMeshCounts(t, first, c.verts, c.indexes)
	}
}

func TestBuiltInMeshDataForPrimitiveKeys(t *testing.T) {
	cases := []PrimitiveMesh{
		PrimitiveMeshSphere,
		PrimitiveMeshTexturableCube,
		PrimitiveMeshCapsule,
		PrimitiveMeshPlane,
		PrimitiveMeshCylinder,
		PrimitiveMeshCone,
		PrimitiveMeshArrow,
	}
	for _, c := range cases {
		cache := NewMeshCache(nil, nil)
		mesh := NewMeshPrimitive(&cache, c)
		verts, indexes, ok := BuiltInMeshData(mesh.Key())
		if !ok {
			t.Fatalf("expected built-in data for %q", mesh.Key())
		}
		if len(verts) != len(mesh.pendingVerts) || len(indexes) != len(mesh.pendingIndexes) {
			t.Fatalf("%q data counts = %d/%d, want %d/%d",
				mesh.Key(), len(verts), len(indexes),
				len(mesh.pendingVerts), len(mesh.pendingIndexes))
		}
	}
	if _, _, ok := BuiltInMeshData("not_a_builtin"); ok {
		t.Fatal("unexpected built-in mesh data for unknown key")
	}
}

func TestMeshAnchoredQuadPivots(t *testing.T) {
	cases := []struct {
		pivot QuadPivot
		key   string
		first matrix.Vec3
	}{
		{QuadPivotCenter, "quad", matrix.Vec3{-0.5, -0.5, 0}},
		{QuadPivotLeft, "quad_left", matrix.Vec3{0, -0.5, 0}},
		{QuadPivotTop, "quad_top", matrix.Vec3{-0.5, -1, 0}},
		{QuadPivotRight, "quad_right", matrix.Vec3{-1, -0.5, 0}},
		{QuadPivotBottom, "quad_bottom", matrix.Vec3{-0.5, 0, 0}},
		{QuadPivotBottomLeft, "quad_bottom_left", matrix.Vec3{0, 0, 0}},
		{QuadPivotBottomRight, "quad_bottom_right", matrix.Vec3{-1, 0, 0}},
		{QuadPivotTopLeft, "quad_top_left", matrix.Vec3{0, -1, 0}},
		{QuadPivotTopRight, "quad_top_right", matrix.Vec3{-1, -1, 0}},
	}
	for _, c := range cases {
		cache := NewMeshCache(nil, nil)
		mesh := NewMeshQuadAnchored(c.pivot, &cache)
		if mesh.Key() != c.key || mesh.pendingVerts[0].Position != c.first {
			t.Fatalf("pivot %d produced %q/%v", c.pivot, mesh.Key(), mesh.pendingVerts[0].Position)
		}
	}
	cache := NewMeshCache(nil, nil)
	if got := NewMeshQuadAnchored(QuadPivot(999), &cache); got.Key() != "quad" {
		t.Fatalf("unknown pivot should default to center, got %q", got.Key())
	}
}

func TestMeshGridLinePointAndWireShapes(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	grid := NewMeshGrid(&cache, "grid", []matrix.Vec3{{0, 0, 0}, {1, 0, 0}}, matrix.ColorRed())
	assertMeshCounts(t, grid, 2, 2)
	point := NewMeshPoint(&cache, "point", matrix.Vec3{1, 2, 3}, matrix.ColorGreen())
	assertMeshCounts(t, point, 1, 1)
	line := NewMeshLine(&cache, "line", matrix.Vec3Zero(), matrix.Vec3One(), matrix.ColorBlue())
	assertMeshCounts(t, line, 2, 2)
	wireQuad := NewMeshWireQuad(&cache, "wire_quad", matrix.ColorWhite())
	assertMeshCounts(t, wireQuad, 4, 8)
	wireCube := NewMeshWireCube(&cache, "wire_cube", matrix.ColorWhite())
	assertMeshCounts(t, wireCube, 8, 24)
	circle := NewMeshCircleWire(&cache, 2, 2)
	assertMeshCounts(t, circle, 3, 6)
	if circle.pendingIndexes[len(circle.pendingIndexes)-1] != 0 {
		t.Fatalf("circle wire should wrap last segment to first vertex")
	}
}

func TestMeshSphereGeneration(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	mesh := NewMeshSphere(&cache, 2, 1, 2)
	assertMeshCounts(t, mesh, (2+1)*(3+1), 2*3*6)
	if mesh.pendingVerts[0].Position != (matrix.Vec3{0, 2, 0}) {
		t.Fatalf("top sphere vertex = %v", mesh.pendingVerts[0].Position)
	}
	if mesh.pendingVerts[0].Normal != matrix.Vec3Up() {
		t.Fatalf("top sphere normal = %v", mesh.pendingVerts[0].Normal)
	}
	if got := NewMeshSphere(&cache, 2, 2, 3); got != mesh {
		t.Fatalf("sphere should reuse clamped cache key")
	}
}

func TestMeshSphereWindingFollowsVertexNormals(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	assertMeshFacesFollowVertexNormals(t, NewMeshSphere(&cache, 1, 8, 8))
}

func TestMeshWireSphereGeneration(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	mesh := NewMeshWireSphere(&cache, 1, 2, 3)
	assertMeshCounts(t, mesh, (4-1)*8, 8*(4-1)*2+8*(4-2)*2)
	for i := 0; i < len(mesh.pendingIndexes); i += 2 {
		if int(mesh.pendingIndexes[i]) >= len(mesh.pendingVerts) ||
			int(mesh.pendingIndexes[i+1]) >= len(mesh.pendingVerts) {
			t.Fatalf("wire sphere index out of range at %d", i)
		}
	}
}

func TestMeshCylinderGeneration(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	uncapped := NewMeshCylinder(&cache, 4, 1, 2, false)
	assertMeshCounts(t, uncapped, 3*2, 3*6)
	if uncapped.pendingVerts[0].Position.Y() != 0 || uncapped.pendingVerts[3].Position.Y() != 4 {
		t.Fatalf("cylinder should be offset into 0..height, got y %v/%v",
			uncapped.pendingVerts[0].Position.Y(), uncapped.pendingVerts[3].Position.Y())
	}
	capped := NewMeshCylinder(&cache, 4, 1, 3, true)
	assertMeshCounts(t, capped, 3*2+3*2, 3*6+3*3*2)
}

func TestMeshConeGeneration(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	uncapped := NewMeshCone(&cache, 3, 1, 2, false)
	assertMeshCounts(t, uncapped, 3+1, 3*3)
	if uncapped.pendingVerts[0].Position != (matrix.Vec3{0, 3, 0}) {
		t.Fatalf("cone apex = %v", uncapped.pendingVerts[0].Position)
	}
	capped := NewMeshCone(&cache, 3, 1, 3, true)
	assertMeshCounts(t, capped, 3+2, 3*3+3*3)
}

func TestMeshArrowGeneration(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	transform := matrix.Mat4Identity()
	transform.Translate(matrix.Vec3{10, 0, 0})
	mesh := NewMeshArrowWithTransform(&cache, 2, 0.25, 1, 0.5, 2, transform, "shifted")
	assertMeshCounts(t, mesh, (3*2+3*2)+(3+2), (3*6+3*3*2)+(3*3+3*3))
	if mesh.pendingVerts[0].Position.X() < 9 {
		t.Fatalf("arrow transform was not applied: %v", mesh.pendingVerts[0].Position)
	}
	if got := NewMeshArrowWithTransform(&cache, 2, 0.25, 1, 0.5, 3, transform, "shifted"); got != mesh {
		t.Fatalf("arrow should reuse cache key")
	}
}

func TestMeshCapsuleGeneration(t *testing.T) {
	cache := NewMeshCache(nil, nil)
	mesh := NewMeshCapsule(&cache, 1, 2, 4, 2)
	assertMeshCounts(t, mesh, (2*2+2)*(4+1), (2*2+1)*4*6)
	for i, index := range mesh.pendingIndexes {
		if int(index) >= len(mesh.pendingVerts) {
			t.Fatalf("capsule index %d = %d, want < %d", i, index, len(mesh.pendingVerts))
		}
	}
	if got := mesh.pendingVerts[0].Position.Y(); got != 2 {
		t.Fatalf("capsule top y = %v, want 2", got)
	}
	if got := mesh.pendingVerts[len(mesh.pendingVerts)-1].Position.Y(); got != -2 {
		t.Fatalf("capsule bottom y = %v, want -2", got)
	}
}
