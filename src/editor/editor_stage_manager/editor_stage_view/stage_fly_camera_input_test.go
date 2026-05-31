/******************************************************************************/
/* stage_fly_camera_input_test.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"testing"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/engine"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/windowing"
)

func TestStageViewFlyCameraInputActiveWhileRightMouseHeldIn3DViewport(t *testing.T) {
	t.Parallel()

	view := stageViewWithTestWindow()
	view.host.Window.Mouse.SetPosition(100, 100, 800, 600)
	view.host.Window.Mouse.SetDown(hid.MouseButtonRight)

	if !view.IsFlyCameraInputActive() {
		t.Fatalf("right mouse in the 3D viewport should capture fly camera input")
	}

	view.host.Window.Keyboard.SetKeyDown(hid.KeyboardKeyLeftAlt)
	if view.IsFlyCameraInputActive() {
		t.Fatalf("alt-right mouse camera gestures should not count as fly camera input")
	}
}

func TestStageViewFlyCameraInputIgnoresRightMouseOutsideViewport(t *testing.T) {
	t.Parallel()

	view := stageViewWithTestWindow()
	view.host.Window.Mouse.SetPosition(900, 100, 800, 600)
	view.host.Window.Mouse.SetDown(hid.MouseButtonRight)

	if view.IsFlyCameraInputActive() {
		t.Fatalf("right mouse outside the viewport should not capture fly camera input")
	}
}

func stageViewWithTestWindow() *StageView {
	win := &windowing.Window{
		Keyboard: hid.NewKeyboard(),
		Mouse:    hid.NewMouse(),
	}
	view := &StageView{
		host: &engine.Host{Window: win},
		open: true,
		viewport: stageViewportBounds{
			Left:   0,
			Top:    0,
			Width:  800,
			Height: 600,
		},
	}
	view.camera.SetViewportBounds(0, 0, 800, 600)
	view.camera.SetModeForRenderView(editor_controls.EditorCameraMode3d, nil)
	return view
}
