/******************************************************************************/
/* light_test.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"testing"

	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
)

func TestLightDirtySetters(t *testing.T) {
	light := Light{
		lightType:   LightTypePoint,
		position:    matrix.Vec3Zero(),
		direction:   matrix.Vec3Down(),
		intensity:   1,
		constant:    1,
		linear:      0.1,
		quadratic:   0.01,
		cutoff:      0.5,
		outerCutoff: 0.75,
		ambient:     matrix.Vec3{0.1, 0.1, 0.1},
		diffuse:     matrix.Vec3One(),
		specular:    matrix.Vec3One(),
	}
	light.SetPosition(matrix.Vec3Zero())
	if light.frameDirty || light.reset {
		t.Fatalf("setting same position should not dirty light")
	}
	light.SetPosition(matrix.Vec3{1, 2, 3})
	if !light.frameDirty || !light.reset {
		t.Fatalf("position change should dirty light")
	}
	light.frameDirty, light.reset = false, false
	light.SetDirection(matrix.Vec3Right())
	light.SetIntensity(2)
	light.SetConstant(3)
	light.SetLinear(4)
	light.SetQuadratic(5)
	light.SetCutoff(0.6)
	light.SetOuterCutoff(0.8)
	light.SetAmbient(matrix.Vec3{0.2, 0.3, 0.4})
	light.SetDiffuse(matrix.Vec3{0.5, 0.6, 0.7})
	light.SetSpecular(matrix.Vec3{0.8, 0.9, 1})
	light.SetCastsShadows(true)
	if !light.frameDirty || !light.reset {
		t.Fatalf("setters should dirty light")
	}
}

func TestLightDirectionalSetPositionIgnored(t *testing.T) {
	light := Light{lightType: LightTypeDirectional, position: matrix.Vec3Zero()}
	light.SetPosition(matrix.Vec3{1, 2, 3})
	if light.position != matrix.Vec3Zero() || light.frameDirty || light.reset {
		t.Fatalf("directional SetPosition should be ignored")
	}
}

func TestLightResetFrameDirty(t *testing.T) {
	light := Light{frameDirty: true}
	if !light.ResetFrameDirty() {
		t.Fatalf("ResetFrameDirty should return previous dirty state")
	}
	if light.frameDirty {
		t.Fatalf("ResetFrameDirty should clear dirty state")
	}
	if light.ResetFrameDirty() {
		t.Fatalf("second ResetFrameDirty should return false")
	}
}

func TestLightTransformToGPU(t *testing.T) {
	camera := cameras.NewStandardCamera(100, 100, 100, 100, matrix.Vec3Zero())
	camera.SetNearPlane(0.25)
	camera.SetFarPlane(50)
	light := Light{
		camera:      camera,
		position:    matrix.Vec3{1, 2, 3},
		direction:   matrix.Vec3Down(),
		intensity:   2,
		cutoff:      0.5,
		outerCutoff: 0.75,
		ambient:     matrix.Vec3{0.1, 0.2, 0.3},
		diffuse:     matrix.Vec3{0.4, 0.5, 0.6},
		specular:    matrix.Vec3{0.7, 0.8, 0.9},
		constant:    1,
		linear:      2,
		quadratic:   3,
		lightType:   LightTypeSpot,
	}
	light.lightSpaceMatrix[0] = matrix.Mat4Identity()
	light.lightSpaceMatrix[0].Translate(matrix.Vec3{9, 0, 0})
	gpu := light.transformToGPULight()
	if gpu.Position != light.position || gpu.Direction != light.direction || gpu.Matrix[0] != light.lightSpaceMatrix[0] {
		t.Fatalf("unexpected GPULight: %+v", gpu)
	}
	info := light.transformToGPULightInfo()
	if info.Position != light.position ||
		info.Direction != light.direction ||
		info.Intensity != light.intensity ||
		info.NearPlane != 0.25 ||
		info.FarPlane != 50 ||
		info.Type != int32(LightTypeSpot) {
		t.Fatalf("unexpected GPULightInfo: %+v", info)
	}
}

func TestLightMinMaxFromCorners(t *testing.T) {
	light := Light{}
	view := matrix.Mat4Identity()
	view.Translate(matrix.Vec3{10, 0, 0})
	corners := graviton.FrustumCorners{
		{-1, -2, -3, 1},
		{4, 5, 6, 1},
		{0, 0, 0, 1},
		{2, -4, 3, 1},
		{-5, 1, 2, 1},
		{3, 2, -1, 1},
		{1, 4, -2, 1},
		{-2, -3, 5, 1},
	}
	mm := light.minMaxFromCorners(view, corners)
	if mm.Min != (matrix.Vec3{5, -4, -3}) || mm.Max != (matrix.Vec3{14, 5, 6}) {
		t.Fatalf("min/max = %+v", mm)
	}
}
