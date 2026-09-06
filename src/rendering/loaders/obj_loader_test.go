/******************************************************************************/
/* obj_loader_test.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package loaders

import (
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

func TestObjDecipherLine(t *testing.T) {
	cases := []struct {
		line string
		typ  objLineType
	}{
		{"# comment", objLineTypeComment},
		{"mtllib foo.mtl", objLineTypeMaterialLib},
		{"o Cube", objLineTypeObject},
		{"v 0 0 0", objLineTypeVertex},
		{"vt 0.5 0.5", objLineTypeUv},
		{"vn 0 1 0", objLineTypeNormal},
		{"usemtl mat1", objLineTypeMaterial},
		{"f 1/1/1 2/2/2 3/3/3", objLineTypeFace},
		{"x unknown", objLineTypeNotSupported},
	}
	for _, c := range cases {
		got := objDecipherLine(c.line)
		if got != c.typ {
			t.Fatalf("objDecipherLine(%q) = %v, want %v", c.line, got, c.typ)
		}
	}
}

func TestObjNewObject(t *testing.T) {
	obj := objNewObject("o MyObject")
	if obj == nil {
		t.Fatalf("objNewObject returned nil")
	}
	if len(obj.points) != 0 {
		t.Fatalf("expected empty points slice, got length %d", len(obj.points))
	}
}

func TestObjLibraryReadVertex(t *testing.T) {
	lib := objLibrary{}
	if err := lib.readVertex("v 1.0 2.0 3.0"); err != nil {
		t.Fatalf("readVertex failed: %v", err)
	}
	if len(lib.points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(lib.points))
	}
	if lib.colors[0] != matrix.ColorWhite() {
		t.Fatalf("expected default white color when not specified")
	}
	// with color components
	lib = objLibrary{}
	if err := lib.readVertex("v 0 0 0 0.1 0.2 0.3"); err != nil {
		t.Fatalf("readVertex with color failed: %v", err)
	}
	if len(lib.colors) != 1 {
		t.Fatalf("expected color slice length 1, got %d", len(lib.colors))
	}
	c := lib.colors[0]
	if *c.PR() != 0.1 || *c.PG() != 0.2 || *c.PB() != 0.3 {
		t.Fatalf("color components not parsed correctly: %+v", c)
	}
}

func TestObjLibraryReadUv(t *testing.T) {
	lib := objLibrary{}
	if err := lib.readUv("vt 0.5 0.25"); err != nil {
		t.Fatalf("readUv failed: %v", err)
	}
	if len(lib.uvs) != 1 {
		t.Fatalf("expected 1 uv, got %d", len(lib.uvs))
	}
}

func TestObjLibraryReadNormal(t *testing.T) {
	lib := objLibrary{}
	if err := lib.readNormal("vn 0 1 0"); err != nil {
		t.Fatalf("readNormal failed: %v", err)
	}
	if len(lib.normals) != 1 {
		t.Fatalf("expected 1 normal, got %d", len(lib.normals))
	}
}

func TestObjLibraryReadMaterial(t *testing.T) {
	lib := objLibrary{}
	if err := lib.readMaterial("mtllib material.mtl"); err != nil {
		t.Fatalf("readMaterial failed: %v", err)
	}
	if len(lib.materials) != 1 || lib.materials[0] != "material.mtl" {
		t.Fatalf("material not stored correctly: %+v", lib.materials)
	}
}

func TestObjBuilderReadFace(t *testing.T) {
	// Setup a library with three vertices, uvs and normals.
	lib := objLibrary{}
	lib.points = []matrix.Vec3{matrix.NewVec3(0, 0, 0), matrix.NewVec3(1, 0, 0), matrix.NewVec3(0, 1, 0)}
	lib.colors = []matrix.Color{matrix.ColorWhite(), matrix.ColorWhite(), matrix.ColorWhite()}
	lib.uvs = []matrix.Vec2{matrix.NewVec2(0, 0), matrix.NewVec2(1, 0), matrix.NewVec2(0, 1)}
	lib.normals = []matrix.Vec3{matrix.NewVec3(0, 0, 1), matrix.NewVec3(0, 0, 1), matrix.NewVec3(0, 0, 1)}

	builder := &objBuilder{name: "test"}
	if err := builder.readFace("f 1/1/1 2/2/2 3/3/3", lib); err != nil {
		t.Fatalf("readFace failed: %v", err)
	}
	if len(builder.points) != 3 {
		t.Fatalf("expected 3 points after face, got %d", len(builder.points))
	}
	if len(builder.vIndexes) != 3 {
		t.Fatalf("expected 3 vIndexes (one triangle), got %d", len(builder.vIndexes))
	}
	if len(builder.tIndexes) != 3 || len(builder.nIndexes) != 3 {
		t.Fatalf("expected matching texture and normal index slices")
	}
}

func TestObjToRawAndOBJ(t *testing.T) {
	objData := `mtllib test.mtl
o Cube
usemtl mat1
v 0 0 0
v 1 0 0
v 0 1 0
vt 0 0
vt 1 0
vt 0 1
vn 0 0 1
vn 0 0 1
vn 0 0 1
f 1/1/1 2/2/2 3/3/3`
	builders, _, err := ObjToRaw(objData)
	if err != nil {
		t.Fatalf("ObjToRaw returned error: %v", err)
	}
	if len(builders) != 1 {
		t.Fatalf("expected 1 builder, got %d", len(builders))
	}
	if builders[0].name != "Cube" {
		t.Fatalf("expected builder name 'Cube', got %s", builders[0].name)
	}
	if len(builders[0].points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(builders[0].points))
	}
	if len(builders[0].vIndexes) != 3 {
		t.Fatalf("expected 3 indexes, got %d", len(builders[0].vIndexes))
	}
	// Test high‑level OBJ loader using mock DB.
	mdb := assets.NewMockDB(map[string][]byte{"model.obj": []byte(objData)})
	result, err := OBJ("model.obj", mdb)
	if err != nil {
		t.Fatalf("OBJ returned error: %v", err)
	}
	if !result.IsValid() {
		t.Fatalf("result should be valid (contain meshes)")
	}
	if len(result.Meshes) != 1 {
		t.Fatalf("expected 1 mesh, got %d", len(result.Meshes))
	}
	// Verify mesh data matches builder data.
	mesh := result.Meshes[0]
	if len(mesh.Verts) != 3 {
		t.Fatalf("mesh should contain 3 vertices, got %d", len(mesh.Verts))
	}
	if len(mesh.Indexes) != 3 {
		t.Fatalf("mesh should contain 3 indexes, got %d", len(mesh.Indexes))
	}
}

func TestObjToRawIgnoresEmptyLines(t *testing.T) {
	objData := `mtllib test.mtl
# vertices
v -176 -16 -64
v -176 -16 144
v -176 0 144
v -176 0 -64
v 48 0 144
v 48 -16 144
v 48 -16 -64
v 48 0 -64

# texture coordinates
vt 1.9999999 -0.5
vt -3.9999995 -0.5
vt -3.9999995 0.5
vt 1.9999999 0.5
vt 1 0.5
vt -6 0.5
vt -6 -0.5
vt 1 -0.5
vt 1 -3.9999995
vt -6 -3.9999995
vt -6 1.9999999
vt 1 1.9999999

# normals
vn -1 0 0
vn 0 0 1
vn 0 -1 -0
vn -0 1 -0
vn 0 -0 -1
vn 1 0 0

o entity0_brush0
usemtl __TB_empty
f  1/1/1  2/2/1  3/3/1  4/4/1
usemtl __TB_empty
f  5/5/2  3/6/2  2/7/2  6/8/2
usemtl __TB_empty
f  6/9/3  2/10/3  1/11/3  7/12/3
usemtl __TB_empty
f  8/12/4  4/11/4  3/10/4  5/9/4
usemtl __TB_empty
f  7/8/5  1/7/5  4/6/5  8/5/5
usemtl __TB_empty
f  8/4/6  5/3/6  6/2/6  7/1/6`

	builders, library, err := ObjToRaw(objData)
	if err != nil {
		t.Fatalf("ObjToRaw returned error: %v", err)
	}
	if len(builders) != 1 {
		t.Fatalf("expected 1 builder, got %d", len(builders))
	}
	if len(library.points) != 8 {
		t.Fatalf("expected 8 points, got %d", len(library.points))
	}
	if len(library.uvs) != 12 {
		t.Fatalf("expected 12 uvs, got %d", len(library.uvs))
	}
	if len(library.normals) != 6 {
		t.Fatalf("expected 6 normals, got %d", len(library.normals))
	}
	if len(builders[0].vIndexes) != 36 {
		t.Fatalf("expected 36 indexes, got %d", len(builders[0].vIndexes))
	}
}
