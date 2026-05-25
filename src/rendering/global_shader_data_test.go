/******************************************************************************/
/* global_shader_data_test.go                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
)

func TestGlobalShaderDataForCameraStoresPerViewData(t *testing.T) {
	uiCamera := cameras.NewStandardCameraOrthographic(800, 600, 800, 600, matrix.Vec3{0, 0, 250})
	leftCamera := cameras.NewStandardCamera(320, 240, 320, 240, matrix.Vec3{-4, 2, -6})
	leftCamera.SetLookAt(matrix.Vec3Zero())
	rightCamera := cameras.NewStandardCamera(640, 240, 640, 240, matrix.Vec3{4, 2, -6})
	rightCamera.SetLookAt(matrix.Vec3Zero())

	left := globalShaderDataForCamera(leftCamera, uiCamera, LightsForRender{}, 1.5, matrix.Vec2{320, 240})
	right := globalShaderDataForCamera(rightCamera, uiCamera, LightsForRender{}, 1.5, matrix.Vec2{640, 240})
	if left.View == right.View {
		t.Fatalf("view matrices should differ between cameras")
	}
	if left.Projection == right.Projection {
		t.Fatalf("projection matrices should differ between view sizes")
	}
	if left.ScreenSize != (matrix.Vec2{320, 240}) || right.ScreenSize != (matrix.Vec2{640, 240}) {
		t.Fatalf("screen sizes were not stored per view: left=%v right=%v", left.ScreenSize, right.ScreenSize)
	}
	if left.CameraPosition == right.CameraPosition {
		t.Fatalf("camera positions should differ between views")
	}
}
