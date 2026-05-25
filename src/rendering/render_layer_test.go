/******************************************************************************/
/* render_layer_test.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "testing"

func TestDrawingDefaultLayerMapsToWorld(t *testing.T) {
	drawing := Drawing{}
	if got := drawing.EffectiveLayer(); got != RenderLayerWorld {
		t.Fatalf("default drawing layer = %v, want world", got)
	}
	if !drawing.MatchesLayer(RenderLayerWorld) {
		t.Fatalf("default drawing should match world layer")
	}
	if drawing.MatchesLayer(RenderLayerUI | RenderLayerEditor) {
		t.Fatalf("default drawing should not match UI/editor layers")
	}
}

func TestDrawingExplicitLayerMatching(t *testing.T) {
	cases := []struct {
		name  string
		layer RenderLayerMask
		mask  RenderLayerMask
		want  bool
	}{
		{name: "world matches world", layer: RenderLayerWorld, mask: RenderLayerWorld, want: true},
		{name: "world skips UI", layer: RenderLayerWorld, mask: RenderLayerUI, want: false},
		{name: "UI matches UI", layer: RenderLayerUI, mask: RenderLayerUI, want: true},
		{name: "UI matches combined mask", layer: RenderLayerUI, mask: RenderLayerWorld | RenderLayerUI, want: true},
		{name: "UI skips editor", layer: RenderLayerUI, mask: RenderLayerEditor, want: false},
		{name: "editor matches editor", layer: RenderLayerEditor, mask: RenderLayerEditor, want: true},
		{name: "editor skips world", layer: RenderLayerEditor, mask: RenderLayerWorld, want: false},
	}
	for _, c := range cases {
		drawing := Drawing{Layer: c.layer}
		if got := drawing.MatchesLayer(c.mask); got != c.want {
			t.Fatalf("%s: MatchesLayer() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestZeroValueDrawingWithMaterialUsesWorldLayer(t *testing.T) {
	drawing := Drawing{Material: &Material{}}
	if !drawing.IsValid() {
		t.Fatalf("drawing with material should remain valid")
	}
	if !drawing.MatchesLayer(RenderLayerWorld) {
		t.Fatalf("zero-value drawing with material should match world layer")
	}

	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{
		Material:   &Material{},
		Mesh:       mesh,
		ShaderData: newTestDrawInstance(),
	})
	if len(drawings.backDraws) != 1 {
		t.Fatalf("expected one pending drawing, got %d", len(drawings.backDraws))
	}
	if got := drawings.backDraws[0].Layer; got != RenderLayerWorld {
		t.Fatalf("queued zero-value drawing layer = %v, want world", got)
	}
}

func TestDrawingsPreparePendingSeparatesLayerGroups(t *testing.T) {
	rp := &RenderPass{}
	mat := &Material{renderPass: rp, Instances: make(map[string]*Material), Textures: []*Texture{{Key: "t"}}}
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance()})
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance(), Layer: RenderLayerUI})
	drawings.PreparePending(0)

	if len(drawings.renderPassGroups) != 1 || len(drawings.renderPassGroups[0].draws) != 1 {
		t.Fatalf("unexpected render grouping: %+v", drawings.renderPassGroups)
	}
	groups := drawings.renderPassGroups[0].draws[0].instanceGroups
	if len(groups) != 2 {
		t.Fatalf("same material/mesh drawings on different layers should not merge, got %d groups", len(groups))
	}
	layerCounts := map[RenderLayerMask]int{}
	for i := range groups {
		layerCounts[groups[i].EffectiveLayer()] += len(groups[i].Instances)
	}
	if layerCounts[RenderLayerWorld] != 1 || layerCounts[RenderLayerUI] != 1 {
		t.Fatalf("unexpected layer grouping: %+v", layerCounts)
	}
}
