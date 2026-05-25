/******************************************************************************/
/* host_render_view_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"testing"

	"kaijuengine.com/rendering"
)

func TestNewHostInitializesDefaultRenderView(t *testing.T) {
	host := NewHost("test", nil, nil)
	view, ok := host.RenderViews.Default()
	if !ok {
		t.Fatalf("default render view was not created")
	}
	if view.Camera() != host.Cameras.Primary.Camera {
		t.Fatalf("default render view camera = %#v, want primary camera", view.Camera())
	}
	if view.LayerMask() != rendering.RenderLayerWorld {
		t.Fatalf("default render view layer mask = %v, want world", view.LayerMask())
	}
	if view.Target() != nil {
		t.Fatalf("default render view should target the swapchain path")
	}
}
