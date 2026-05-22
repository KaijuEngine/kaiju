/******************************************************************************/
/* material_drawing_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"
	"weak"
)

func TestMaterialHasTransparentSuffix(t *testing.T) {
	if !(&Material{Id: "ui_transparent"}).HasTransparentSuffix() {
		t.Fatalf("expected transparent suffix")
	}
	if (&Material{Id: "transparent_ui"}).HasTransparentSuffix() {
		t.Fatalf("suffix check should require _transparent")
	}
}

func TestMaterialSelectRoot(t *testing.T) {
	root := &Material{Id: "root"}
	child := &Material{Id: "child", Root: weak.Make(root)}
	if child.SelectRoot() != root {
		t.Fatalf("child should select root")
	}
	if root.SelectRoot() != root {
		t.Fatalf("root should select itself")
	}
}

func TestMaterialCreateInstance(t *testing.T) {
	root := &Material{Id: "root", Instances: make(map[string]*Material)}
	textures := []*Texture{{Key: "a"}, {Key: "b"}}
	inst := root.CreateInstance(textures)
	if inst == root {
		t.Fatalf("CreateInstance should create a copy")
	}
	if inst.Root.Value() != root {
		t.Fatalf("instance root was not set")
	}
	if inst.Instances != nil {
		t.Fatalf("instances should not keep their own cache map")
	}
	if len(inst.Textures) != len(textures) || inst.Textures[0] != textures[0] {
		t.Fatalf("textures were not copied")
	}
	textures[0] = &Texture{Key: "changed"}
	if inst.Textures[0].Key != "a" {
		t.Fatalf("texture slice should be cloned")
	}
	if got := root.CreateInstance([]*Texture{{Key: "a"}, {Key: "b"}}); got != inst {
		t.Fatalf("same texture keys should reuse instance")
	}
}

func TestMaterialCreateInstanceWithPrepass(t *testing.T) {
	prepass := &Material{Id: "prepass", Instances: make(map[string]*Material)}
	root := &Material{Id: "root", Instances: make(map[string]*Material), PrepassMaterial: weak.Make(prepass)}
	tex := []*Texture{{Key: "a"}}
	inst := root.CreateInstance(tex)
	if inst.PrepassMaterial.Value() == nil {
		t.Fatalf("prepass instance was not created")
	}
	if inst.PrepassMaterial.Value() == prepass {
		t.Fatalf("prepass should point at a matching instance, not the root prepass")
	}
	if inst.PrepassMaterial.Value().Root.Value() != prepass {
		t.Fatalf("prepass instance root was not set")
	}
}

func TestMaterialTextureDataFilterToVK(t *testing.T) {
	cases := []struct {
		filter string
		want   TextureFilter
	}{
		{"Nearest", TextureFilterNearest},
		{"Linear", TextureFilterLinear},
		{"CubicImg", TextureFilterLinear},
		{"", TextureFilterLinear},
		{"bad", TextureFilterLinear},
	}
	for _, c := range cases {
		if got := (&MaterialTextureData{Filter: c.filter}).FilterToVK(); got != c.want {
			t.Fatalf("FilterToVK(%q) = %v, want %v", c.filter, got, c.want)
		}
	}
}

func TestDrawingIsValid(t *testing.T) {
	if !(&Drawing{Material: &Material{}}).IsValid() {
		t.Fatalf("drawing with material should be valid")
	}
	if (&Drawing{}).IsValid() {
		t.Fatalf("drawing without material should be invalid")
	}
}

func TestNewDrawings(t *testing.T) {
	drawings := NewDrawings()
	if drawings.HasDrawings() || len(drawings.backDraws) != 0 || len(drawings.renderPassGroups) != 0 {
		t.Fatalf("new drawings should be empty: %+v", drawings)
	}
}

func TestDrawingsAddDrawing(t *testing.T) {
	prepass := &Material{}
	mat := &Material{PrepassMaterial: weak.Make(prepass)}
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance()})
	if len(drawings.backDraws) != 2 {
		t.Fatalf("prepass should add two pending drawings, got %d", len(drawings.backDraws))
	}
	if drawings.backDraws[0].Material != prepass || drawings.backDraws[1].Material != mat {
		t.Fatalf("unexpected pending materials")
	}
	assertPanics(t, func() {
		d := NewDrawings()
		d.AddDrawing(Drawing{Material: mat, ShaderData: newTestDrawInstance()})
	})
	assertPanics(t, func() {
		d := NewDrawings()
		d.AddDrawing(Drawing{Mesh: mesh, ShaderData: newTestDrawInstance()})
	})
}

func TestDrawingsPreparePendingGroups(t *testing.T) {
	rp := &RenderPass{}
	mat := &Material{renderPass: rp, Instances: make(map[string]*Material), Textures: []*Texture{{Key: "t"}}}
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance(), Sort: 5})
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance(), Sort: 5})
	drawings.PreparePending(0)
	if len(drawings.backDraws) != 0 {
		t.Fatalf("pending drawings were not cleared")
	}
	if !drawings.HasDrawings() || len(drawings.renderPassGroups) != 1 {
		t.Fatalf("render pass group was not created")
	}
	group := drawings.renderPassGroups[0]
	if group.renderPass != rp || len(group.draws) != 1 || len(group.draws[0].instanceGroups) != 1 {
		t.Fatalf("unexpected grouping: %+v", group)
	}
	instances := group.draws[0].instanceGroups[0]
	if instances.Mesh != mesh || len(instances.Instances) != 2 || instances.sort != 5 ||
		instances.MaterialInstance.Textures[0].Key != "t" {
		t.Fatalf("unexpected instance group: %+v", instances)
	}
}

func TestDrawingsClear(t *testing.T) {
	rp := &RenderPass{}
	mat := &Material{renderPass: rp, Instances: make(map[string]*Material)}
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	inst := newTestDrawInstance()
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: inst})
	drawings.PreparePending(0)
	drawings.Clear()
	if !inst.IsDestroyed() {
		t.Fatalf("Clear should destroy draw instances")
	}
	if len(drawings.renderPassGroups) != 1 {
		t.Fatalf("Clear should keep render pass groups")
	}
}

func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()
	fn()
}
