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

func TestGlobalShaderDataAssignsOnlyOneDirectionalShadowSlot(t *testing.T) {
	camera := cameras.NewStandardCamera(320, 240, 320, 240, matrix.Vec3{0, 2, -6})
	device := &GPUDevice{}
	first := NewLight(device, nil, nil, LightTypeDirectional)
	first.SetCastsShadows(true)
	second := NewLight(device, nil, nil, LightTypeDirectional)
	second.SetCastsShadows(true)
	point := NewLight(device, nil, nil, LightTypePoint)
	point.SetCastsShadows(true)

	data := globalShaderDataForCamera(camera, nil, LightsForRender{
		Lights: []Light{first, second, point},
	}, 0, matrix.Vec2{320, 240})
	if data.LightInfos[0].ShadowIndex != 0 {
		t.Fatalf("first directional shadow index = %d, want 0", data.LightInfos[0].ShadowIndex)
	}
	if data.LightInfos[1].ShadowIndex != -1 || data.LightInfos[2].ShadowIndex != -1 {
		t.Fatalf("non-selected shadow indexes = %d, %d; want -1, -1",
			data.LightInfos[1].ShadowIndex, data.LightInfos[2].ShadowIndex)
	}

	first.SetCastsShadows(false)
	second.SetCastsShadows(false)
	data = globalShaderDataForCamera(camera, nil, LightsForRender{
		Lights: []Light{first, second, point},
	}, 0, matrix.Vec2{320, 240})
	for i := range 3 {
		if data.LightInfos[i].ShadowIndex != -1 {
			t.Fatalf("disabled light %d shadow index = %d, want -1", i, data.LightInfos[i].ShadowIndex)
		}
	}
}
