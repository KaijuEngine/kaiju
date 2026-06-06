/******************************************************************************/
/* stage_picking_test.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/rendering"
)

func TestStagePickingDisableKeepsPersistentRenderResources(t *testing.T) {
	view := stageViewWithRenderManagers()
	target, err := view.host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   stagePickingRenderName,
		Width:  16,
		Height: 16,
	})
	if err != nil {
		t.Fatalf("Create target returned error: %v", err)
	}
	renderView, err := view.host.RenderViews.Create(rendering.RenderViewOptions{
		Name:   stagePickingRenderName,
		Target: target,
		Clear:  true,
	})
	if err != nil {
		t.Fatalf("Create render view returned error: %v", err)
	}
	picking := StagePicking{
		view:       view,
		target:     target,
		renderView: renderView,
	}

	picking.disableRenderView()

	if picking.target != target || picking.renderView != renderView {
		t.Fatalf("disableRenderView should keep persistent target/view")
	}
	if renderView.Enabled() {
		t.Fatalf("disableRenderView should disable the render view")
	}
	if _, ok := view.host.RenderTargets.Target(stagePickingRenderName); !ok {
		t.Fatalf("disableRenderView removed the render target")
	}
	if _, ok := view.host.RenderViews.View(stagePickingRenderName); !ok {
		t.Fatalf("disableRenderView removed the render view")
	}
}

func TestStagePickingCloseDestroysPersistentRenderResources(t *testing.T) {
	view := stageViewWithRenderManagers()
	target, err := view.host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   stagePickingRenderName,
		Width:  16,
		Height: 16,
	})
	if err != nil {
		t.Fatalf("Create target returned error: %v", err)
	}
	renderView, err := view.host.RenderViews.Create(rendering.RenderViewOptions{
		Name:   stagePickingRenderName,
		Target: target,
		Clear:  true,
	})
	if err != nil {
		t.Fatalf("Create render view returned error: %v", err)
	}
	picking := StagePicking{
		view:       view,
		target:     target,
		renderView: renderView,
	}

	picking.Close()

	if picking.target != nil || picking.renderView != nil {
		t.Fatalf("Close should clear persistent target/view references")
	}
	if _, ok := view.host.RenderTargets.Target(stagePickingRenderName); ok {
		t.Fatalf("Close left the render target in the manager")
	}
	if _, ok := view.host.RenderViews.View(stagePickingRenderName); ok {
		t.Fatalf("Close left the render view in the manager")
	}
}

func stageViewWithRenderManagers() *StageView {
	view := stageViewWithTestWindow()
	view.host.RenderTargets = rendering.NewRenderTargetManager()
	view.host.RenderViews = rendering.NewRenderViewManager(rendering.RenderViewOptions{
		Name:      rendering.DefaultRenderViewName,
		LayerMask: rendering.RenderLayerAll,
		Clear:     true,
	})
	return view
}
