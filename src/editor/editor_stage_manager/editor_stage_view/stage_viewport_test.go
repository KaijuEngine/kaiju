/******************************************************************************/
/* stage_viewport_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestStageViewportBoundsConvertsScreenToLocalCoordinates(t *testing.T) {
	t.Parallel()

	bounds := stageViewportBounds{Left: 100, Top: 50, Width: 400, Height: 300}
	screen := matrix.NewVec2(150, 80)

	if got := bounds.LocalTopFromScreen(screen); got != matrix.NewVec2(50, 30) {
		t.Fatalf("local top position = %v, want %v", got, matrix.NewVec2(50, 30))
	}
	if got := bounds.LocalBottomFromScreen(screen); got != matrix.NewVec2(50, 270) {
		t.Fatalf("local bottom position = %v, want %v", got, matrix.NewVec2(50, 270))
	}
	if !bounds.ContainsScreenPosition(screen) {
		t.Fatal("expected screen position inside viewport")
	}
	if bounds.ContainsScreenPosition(matrix.NewVec2(99, 80)) {
		t.Fatal("expected screen position outside viewport")
	}
}

func TestStageViewportBoundsConvertsScreenBoxToLocalBottomArea(t *testing.T) {
	t.Parallel()

	bounds := stageViewportBounds{Left: 100, Top: 50, Width: 400, Height: 300}
	box := matrix.NewVec4(150, 80, 250, 180)

	got := bounds.LocalBottomAreaFromScreenArea(box)
	want := matrix.NewVec4(50, 170, 150, 270)
	if got != want {
		t.Fatalf("local bottom area = %v, want %v", got, want)
	}
}

func TestStageTargetResizeFollowsViewportPanelSize(t *testing.T) {
	t.Parallel()

	manager := rendering.NewRenderTargetManager()
	target, err := manager.Create(rendering.RenderTargetOptions{
		Name:   "stage-main-test",
		Width:  10,
		Height: 10,
		Depth:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !resizeStageTargetToViewport(target, matrix.NewVec2(320.2, 180.1)) {
		t.Fatal("expected target resize")
	}
	if gotW, gotH := target.Size(); gotW != 321 || gotH != 181 {
		t.Fatalf("target size = %dx%d, want 321x181", gotW, gotH)
	}
	if resizeStageTargetToViewport(target, matrix.NewVec2(321, 181)) {
		t.Fatal("did not expect resize when panel size already matches target")
	}
}

func TestStageViewportRoutingUsesHoveredAndFocusedViewport(t *testing.T) {
	t.Parallel()

	viewports := []stageRenderViewport{
		{Kind: StageViewportPerspective, bounds: stageViewportBounds{Left: 0, Top: 0, Width: 100, Height: 100}},
		{Kind: StageViewportTop, bounds: stageViewportBounds{Left: 100, Top: 0, Width: 100, Height: 100}},
	}

	active, focused, hovered := resolveStageViewportRouting(
		viewports, 0, -1, matrix.NewVec2(125, 50), true, false, false)
	if active != 1 || focused != 1 || hovered != 1 {
		t.Fatalf("pressed routing = active:%d focused:%d hovered:%d, want 1/1/1",
			active, focused, hovered)
	}

	active, focused, hovered = resolveStageViewportRouting(
		viewports, active, focused, matrix.NewVec2(20, 50), false, true, false)
	if active != 1 || focused != 1 || hovered != 0 {
		t.Fatalf("held routing = active:%d focused:%d hovered:%d, want 1/1/0",
			active, focused, hovered)
	}

	active, focused, hovered = resolveStageViewportRouting(
		viewports, active, focused, matrix.NewVec2(20, 50), false, false, true)
	if active != 1 || focused != -1 || hovered != 0 {
		t.Fatalf("released routing = active:%d focused:%d hovered:%d, want 1/-1/0",
			active, focused, hovered)
	}

	active, focused, hovered = resolveStageViewportRouting(
		viewports, active, focused, matrix.NewVec2(20, 50), false, false, false)
	if active != 0 || focused != -1 || hovered != 0 {
		t.Fatalf("post-release routing = active:%d focused:%d hovered:%d, want 0/-1/0",
			active, focused, hovered)
	}
}

func TestStageViewportRoutingIgnoresInvalidStaleBounds(t *testing.T) {
	t.Parallel()

	viewports := []stageRenderViewport{
		{Kind: StageViewportPerspective, bounds: stageViewportBounds{}},
		{Kind: StageViewportTop, bounds: stageViewportBounds{Left: 0, Top: 0, Width: 100, Height: 100}},
	}

	active, focused, hovered := resolveStageViewportRouting(
		viewports, 0, -1, matrix.NewVec2(50, 50), false, false, false)
	if active != 1 || focused != -1 || hovered != 1 {
		t.Fatalf("routing = active:%d focused:%d hovered:%d, want 1/-1/1",
			active, focused, hovered)
	}
}

func TestStageViewportRoutingIgnoresInactiveViewportWithStaleBounds(t *testing.T) {
	t.Parallel()

	hiddenPerspectiveUI := &ui.UI{}
	activeTopUI := &ui.UI{}
	activeTopUI.Show()
	viewports := []stageRenderViewport{
		{
			Kind:   StageViewportPerspective,
			ui:     hiddenPerspectiveUI,
			bounds: stageViewportBounds{Left: 0, Top: 0, Width: 100, Height: 100},
		},
		{
			Kind:   StageViewportTop,
			ui:     activeTopUI,
			bounds: stageViewportBounds{Left: 0, Top: 0, Width: 200, Height: 200},
		},
	}

	active, focused, hovered := resolveStageViewportRouting(
		viewports, 0, -1, matrix.NewVec2(25, 25), true, false, false)
	if active != 1 || focused != 1 || hovered != 1 {
		t.Fatalf("routing = active:%d focused:%d hovered:%d, want 1/1/1",
			active, focused, hovered)
	}
}

func TestStageViewportsOwnDistinctCamerasAndTargets(t *testing.T) {
	t.Parallel()

	host := &engine.Host{
		RenderTargets: rendering.NewRenderTargetManager(),
		RenderViews:   rendering.NewRenderViewManager(),
	}
	view := StageView{host: host}
	for _, kind := range StageViewportKinds() {
		camera := &editor_controls.EditorCamera{}
		camera.SetViewportBounds(0, 0, 320, 180)
		camera.SetModeForRenderView(kind.cameraMode(), host)
		view.stageViewports = append(view.stageViewports, stageRenderViewport{
			Kind:   kind,
			Label:  kind.Label(),
			camera: camera,
			bounds: stageViewportBounds{Width: 320, Height: 180},
		})
	}

	seenCameras := make(map[any]StageViewportKind, len(view.stageViewports))
	seenTargets := make(map[*rendering.RenderTarget]StageViewportKind, len(view.stageViewports))
	for i := range view.stageViewports {
		viewport := &view.stageViewports[i]
		view.ensureStageRenderTarget(viewport)
		if viewport.camera == nil || viewport.camera.Camera() == nil {
			t.Fatalf("%s viewport has no camera", viewport.Label)
		}
		if previous, ok := seenCameras[viewport.camera.Camera()]; ok {
			t.Fatalf("%s and %s share a camera", previous.Label(), viewport.Kind.Label())
		}
		seenCameras[viewport.camera.Camera()] = viewport.Kind
		if viewport.target == nil {
			t.Fatalf("%s viewport has no render target", viewport.Label)
		}
		if previous, ok := seenTargets[viewport.target]; ok {
			t.Fatalf("%s and %s share a render target", previous.Label(), viewport.Kind.Label())
		}
		seenTargets[viewport.target] = viewport.Kind
		if viewport.renderView == nil {
			t.Fatalf("%s viewport has no render view", viewport.Label)
		}
		if viewport.renderView.LayerMask() != (rendering.RenderLayerWorld | rendering.RenderLayerEditor) {
			t.Fatalf("%s layer mask = %d, want world|editor", viewport.Label, viewport.renderView.LayerMask())
		}
	}
}

func TestStageViewOpenUsesUIOnlyDefaultViewAndCloseRestores(t *testing.T) {
	t.Parallel()

	host := &engine.Host{
		RenderViews: rendering.NewRenderViewManager(rendering.RenderViewOptions{
			Name:      rendering.DefaultRenderViewName,
			LayerMask: rendering.RenderLayerAll,
			Clear:     true,
			Sort:      7,
			ViewMode:  rendering.RenderViewModeWireframe,
		}),
	}
	view := StageView{host: host}

	view.Open()

	defaultView, ok := host.RenderViews.Default()
	if !ok {
		t.Fatal("default render view missing after stage open")
	}
	if defaultView.LayerMask() != rendering.RenderLayerUI {
		t.Fatalf("stage default layer mask = %v, want UI only", defaultView.LayerMask())
	}
	if defaultView.Sort() != 7 || defaultView.ViewMode() != rendering.RenderViewModeWireframe {
		t.Fatalf("stage default view did not preserve sort/mode")
	}

	view.Close()

	defaultView, ok = host.RenderViews.Default()
	if !ok {
		t.Fatal("default render view missing after stage close")
	}
	if defaultView.LayerMask() != rendering.RenderLayerAll {
		t.Fatalf("restored default layer mask = %v, want all", defaultView.LayerMask())
	}
	if defaultView.Sort() != 7 || defaultView.ViewMode() != rendering.RenderViewModeWireframe {
		t.Fatalf("restored default view did not preserve sort/mode")
	}
}
