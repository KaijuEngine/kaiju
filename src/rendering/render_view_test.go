/******************************************************************************/
/* render_view_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "testing"

func TestRenderViewManagerSortsDeterministically(t *testing.T) {
	manager := NewRenderViewManager(RenderViewOptions{
		Name:      DefaultRenderViewName,
		LayerMask: RenderLayerWorld,
		Sort:      10,
	})
	mustCreateRenderView(t, &manager, RenderViewOptions{Name: "same-b", Sort: 2})
	mustCreateRenderView(t, &manager, RenderViewOptions{Name: "first", Sort: -1})
	mustCreateRenderView(t, &manager, RenderViewOptions{Name: "same-a", Sort: 2})

	views := manager.Views()
	names := make([]string, len(views))
	for i := range views {
		names[i] = views[i].Name()
	}
	want := []string{"first", "same-a", "same-b", DefaultRenderViewName}
	if len(names) != len(want) {
		t.Fatalf("view count = %d, want %d: %v", len(names), len(want), names)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("view order = %v, want %v", names, want)
		}
	}
}

func TestRenderViewLayerMaskSelectsMatchingDrawingGroups(t *testing.T) {
	rp := &RenderPass{}
	mat := &Material{renderPass: rp, Instances: make(map[string]*Material), Textures: []*Texture{{Key: "t"}}}
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance()})
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance(), Layer: RenderLayerUI})
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance(), Layer: RenderLayerEditor})
	drawings.PreparePending(0)

	view := newRenderView(RenderViewOptions{
		Name:      "editor-world",
		LayerMask: RenderLayerWorld | RenderLayerEditor,
	}, 0)
	groups := drawings.renderPassGroups[0].draws[0].Filter(view.MatchesGroup)
	if len(groups) != 2 {
		t.Fatalf("selected group count = %d, want 2", len(groups))
	}
	counts := map[RenderLayerMask]int{}
	for i := range groups {
		counts[groups[i].EffectiveLayer()] += len(groups[i].Instances)
	}
	if counts[RenderLayerWorld] != 1 || counts[RenderLayerEditor] != 1 {
		t.Fatalf("selected layer counts = %+v, want world/editor", counts)
	}
	if counts[RenderLayerUI] != 0 {
		t.Fatalf("UI layer should not be selected: %+v", counts)
	}
}

func TestRenderViewManagerCreatesImplicitDefaultView(t *testing.T) {
	manager := NewRenderViewManager()
	view, ok := manager.Default()
	if !ok {
		t.Fatalf("implicit default render view was not created")
	}
	if view.Name() != DefaultRenderViewName {
		t.Fatalf("default render view name = %q, want %q", view.Name(), DefaultRenderViewName)
	}
	if view.LayerMask() != RenderLayerWorld {
		t.Fatalf("default layer mask = %v, want world", view.LayerMask())
	}
	if !view.Clear() {
		t.Fatalf("default render view should clear")
	}
}

func TestRenderViewsForDrawKeepsDefaultViewActive(t *testing.T) {
	manager := NewRenderViewManager(RenderViewOptions{
		Name:      DefaultRenderViewName,
		LayerMask: RenderLayerWorld,
	})
	other := mustCreateRenderView(t, &manager, RenderViewOptions{Name: "offscreen"})
	defaultView, _ := manager.Default()
	selected := renderViewsForDraw([]*RenderView{other, defaultView})
	if len(selected) != 1 || selected[0] != defaultView {
		t.Fatalf("selected views = %v, want only default view", selected)
	}
}

func mustCreateRenderView(t *testing.T, manager *RenderViewManager, options RenderViewOptions) *RenderView {
	t.Helper()
	view, err := manager.Create(options)
	if err != nil {
		t.Fatalf("Create(%q) returned error: %v", options.Name, err)
	}
	return view
}
