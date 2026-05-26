/******************************************************************************/
/* editor_camera_test.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_controls

import (
	"testing"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

func TestEditorCameraFocusPreservesFixedOrthographicView(t *testing.T) {
	t.Parallel()

	for _, mode := range []EditorCameraMode{
		EditorCameraModeTop,
		EditorCameraModeFront,
		EditorCameraModeSide,
	} {
		t.Run(cameraModeStrings[mode], func(t *testing.T) {
			t.Parallel()

			editorCamera := &EditorCamera{}
			editorCamera.SetViewportBounds(0, 0, 800, 600)
			editorCamera.SetModeForRenderView(mode, nil)

			camera := editorCamera.Camera().(*cameras.StandardCamera)
			forward := camera.Forward()
			up := camera.Up()
			offset := camera.Position().Subtract(camera.LookAt())
			bounds := graviton.NewAABB(
				matrix.NewVec3(3, 4, 5),
				matrix.NewVec3(1, 2, 3),
			)

			editorCamera.Focus(bounds)

			if !matrix.Vec3ApproxTo(camera.LookAt(), bounds.Center, 0.0001) {
				t.Fatalf("look at = %v, want %v", camera.LookAt(), bounds.Center)
			}
			if got := camera.Position().Subtract(camera.LookAt()); !matrix.Vec3ApproxTo(got, offset, 0.0001) {
				t.Fatalf("position offset = %v, want %v", got, offset)
			}
			if !matrix.Vec3ApproxTo(camera.Forward(), forward, 0.0001) {
				t.Fatalf("forward = %v, want %v", camera.Forward(), forward)
			}
			if !matrix.Vec3ApproxTo(camera.Up(), up, 0.0001) {
				t.Fatalf("up = %v, want %v", camera.Up(), up)
			}
		})
	}
}

func TestEditorCameraPanMovesFixedOrthographicCamera(t *testing.T) {
	t.Parallel()

	for _, mode := range []EditorCameraMode{
		EditorCameraModeTop,
		EditorCameraModeFront,
		EditorCameraModeSide,
	} {
		t.Run(cameraModeStrings[mode], func(t *testing.T) {
			t.Parallel()

			editorCamera := &EditorCamera{}
			editorCamera.SetViewportBounds(0, 0, 800, 600)
			editorCamera.SetModeForRenderView(mode, nil)

			camera := editorCamera.Camera().(*cameras.StandardCamera)
			from := matrix.NewVec2(400, 300)
			to := matrix.NewVec2(450, 330)
			beforePosition := camera.Position()
			beforeLookAt := camera.LookAt()
			beforeForward := camera.Forward()
			beforeUp := camera.Up()
			dx := (from.X() - to.X()) * camera.Width() / 800
			dy := (from.Y() - to.Y()) * camera.Height() / 600
			wantDelta := camera.Right().Scale(dx).Add(camera.Up().Scale(dy))

			editorCamera.panFixedOrthographic(camera, from, to, nil)

			if !matrix.Vec3ApproxTo(camera.LookAt(), beforeLookAt.Add(wantDelta), 0.0001) {
				t.Fatalf("look at = %v, want %v", camera.LookAt(), beforeLookAt.Add(wantDelta))
			}
			if !matrix.Vec3ApproxTo(camera.Position(), beforePosition.Add(wantDelta), 0.0001) {
				t.Fatalf("position = %v, want %v", camera.Position(), beforePosition.Add(wantDelta))
			}
			if !matrix.Vec3ApproxTo(camera.Forward(), beforeForward, 0.0001) {
				t.Fatalf("forward = %v, want %v", camera.Forward(), beforeForward)
			}
			if !matrix.Vec3ApproxTo(camera.Up(), beforeUp, 0.0001) {
				t.Fatalf("up = %v, want %v", camera.Up(), beforeUp)
			}
		})
	}
}
