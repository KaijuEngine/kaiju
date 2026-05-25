/******************************************************************************/
/* global_shader_data.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/matrix"
)

const (
	MaxJoints        = 50
	MaxSkinInstances = 50
)

type PointShadow struct {
	Point    matrix.Vec2 // X,Z
	Radius   float32
	Strength float32
}

type GlobalShaderData struct {
	View                  matrix.Mat4
	Projection            matrix.Mat4
	UIView                matrix.Mat4
	UIProjection          matrix.Mat4
	CameraPosition        matrix.Vec4
	UICameraPosition      matrix.Vec3
	Time                  float32
	ScreenSize            matrix.Vec2
	CascadeCount          int32
	_                     float32
	CascadePlaneDistances [4]float32
	VertLights            [MaxLocalLights]GPULight
	LightInfos            [MaxLocalLights]GPULightInfo
}

func globalShaderDataForCamera(camera cameras.Camera, uiCamera cameras.Camera, lights LightsForRender, runtime float32, screenSize matrix.Vec2) GlobalShaderData {
	ubo := GlobalShaderData{
		View:         matrix.Mat4Identity(),
		Projection:   matrix.Mat4Identity(),
		UIView:       matrix.Mat4Identity(),
		UIProjection: matrix.Mat4Identity(),
		Time:         runtime,
		ScreenSize:   screenSize,
	}
	if camera != nil {
		camOrtho := matrix.Float(0)
		if camera.IsOrthographic() {
			camOrtho = 1
		}
		ubo.View = camera.View()
		ubo.Projection = camera.Projection()
		ubo.CameraPosition = camera.Position().AsVec4WithW(camOrtho)
		ubo.CascadeCount = int32(camera.NumCSMCascades())
		ubo.CascadePlaneDistances = camera.CSMCascadeDistances()
	}
	if uiCamera != nil {
		ubo.UIView = uiCamera.View()
		ubo.UIProjection = uiCamera.Projection()
		ubo.UICameraPosition = uiCamera.Position()
	}
	if camera != nil {
		for i := range lights.Lights {
			if lights.Lights[i].IsValid() {
				lights.Lights[i].recalculate(camera)
				ubo.VertLights[i] = lights.Lights[i].transformToGPULight()
				ubo.LightInfos[i] = lights.Lights[i].transformToGPULightInfo()
			}
		}
	}
	return ubo
}
