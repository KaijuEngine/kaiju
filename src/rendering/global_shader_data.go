/******************************************************************************/
/* global_shader_data.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import "kaijuengine.com/matrix"

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
