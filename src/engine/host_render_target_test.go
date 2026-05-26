/******************************************************************************/
/* host_render_target_test.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"testing"

	"kaijuengine.com/rendering"
)

func TestNewHostInitializesRenderTargetManager(t *testing.T) {
	host := NewHost("test", nil, nil)
	target, err := host.RenderTargets.Create(rendering.RenderTargetOptions{
		Name:   "game",
		Width:  640,
		Height: 360,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got, ok := host.RenderTargets.Target("game"); !ok || got != target {
		t.Fatalf("Host render target manager did not return created target")
	}
}
