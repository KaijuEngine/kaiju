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
	if view.LayerMask() != RenderLayerAll {
		t.Fatalf("default layer mask = %v, want all", view.LayerMask())
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

func TestRenderViewsForDrawPlacesTargetsBeforeDefault(t *testing.T) {
	target := mustCreateRenderTarget(t, RenderTargetOptions{
		Name:   "viewport",
		Width:  320,
		Height: 200,
	})
	manager := NewRenderViewManager(RenderViewOptions{
		Name:      DefaultRenderViewName,
		LayerMask: RenderLayerAll,
	})
	targetView := mustCreateRenderView(t, &manager, RenderViewOptions{
		Name:      "viewport",
		Target:    target,
		LayerMask: RenderLayerWorld,
		Sort:      10,
	})
	defaultView, _ := manager.Default()
	selected := renderViewsForDraw(manager.Views())
	if len(selected) != 2 {
		t.Fatalf("selected view count = %d, want 2", len(selected))
	}
	if selected[0] != targetView || selected[1] != defaultView {
		t.Fatalf("selected view order = %v, want target then default", selected)
	}
}

func TestRenderViewDestroyMarksViewDestroyedAndRemovesIt(t *testing.T) {
	manager := NewRenderViewManager()
	view := mustCreateRenderView(t, &manager, RenderViewOptions{Name: "preview"})
	if err := manager.Destroy("preview"); err != nil {
		t.Fatalf("Destroy returned error: %v", err)
	}
	if !view.Destroyed() {
		t.Fatalf("destroyed view was not marked destroyed")
	}
	if _, ok := manager.View("preview"); ok {
		t.Fatalf("destroyed view remained in manager lookup")
	}
	for _, candidate := range manager.Views() {
		if candidate == view {
			t.Fatalf("destroyed view remained in sorted view list")
		}
	}
}

func TestRenderViewProcessPendingDestroysDrawGroupViewState(t *testing.T) {
	manager := NewRenderViewManager()
	view := mustCreateRenderView(t, &manager, RenderViewOptions{Name: "preview"})
	rp := &RenderPass{}
	mat := &Material{renderPass: rp, Instances: make(map[string]*Material), Textures: []*Texture{{Key: "t"}}}
	mesh := NewMesh("mesh", testVerts(), []uint32{0, 1})
	drawings := NewDrawings()
	drawings.AddDrawing(Drawing{Material: mat, Mesh: mesh, ShaderData: newTestDrawInstance()})
	drawings.PreparePending(0)

	group := &drawings.renderPassGroups[0].draws[0].instanceGroups[0]
	group.viewStateForView(view)
	if _, ok := group.viewStates[view]; !ok {
		t.Fatalf("test setup did not create view state")
	}
	if err := manager.Destroy("preview"); err != nil {
		t.Fatalf("Destroy returned error: %v", err)
	}
	manager.ProcessPending(nil, &drawings)
	if _, ok := group.viewStates[view]; ok {
		t.Fatalf("pending view cleanup left draw group view state behind")
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
