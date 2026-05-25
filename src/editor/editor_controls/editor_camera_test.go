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
