/******************************************************************************/
/* light.go                                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rendering

import (
	"kaiju/engine/assets"
	"kaiju/engine/cameras"
	"kaiju/matrix"
	vk "kaiju/rendering/vulkan"
	"log/slog"
	"unsafe"
	"weak"
)

const (
	nrLights                 = 4
	maxLights                = 20
	cubeMapSides             = 6
	lightDepthMapWidth       = 4096
	lightDepthMapHeight      = 4096
	lightDirectionalScaleOut = 50.0
	lightShadowmapFilter     = vk.FilterLinear
	lightDepthFormat         = vk.FormatD16Unorm
	//lightDepthFormat       = vk.FormatD32Sfloat
)

var (
	lightDepthMaterial     weak.Pointer[Material]
	lightCubeDepthMaterial weak.Pointer[Material]
)

type LightType int

const (
	LightTypeDirectional = LightType(iota)
	LightTypePoint
	LightTypeSpot
)

type GPULight struct {
	Matrix    [cubeMapSides]matrix.Mat4
	Position  matrix.Vec3
	_         float32
	Direction matrix.Vec3
	_         float32
}

type GPULightInfo struct {
	Position    matrix.Vec3
	Intensity   float32
	Direction   matrix.Vec3
	Cutoff      float32
	Ambient     matrix.Vec3
	OuterCutoff float32
	Diffuse     matrix.Vec3
	Constant    float32
	Specular    matrix.Vec3
	Linear      float32
	Quadratic   float32
	NearPlane   float32
	FarPlane    float32
	Type        int32
	_           float32
}

type Light struct {
	renderer         *Vulkan
	depthMaterial    *Material
	camera           cameras.Camera
	renderPass       *RenderPass
	vulkanDepthMap   TextureId
	lightSpaceMatrix [cubeMapSides]matrix.Mat4
	ambient          matrix.Vec3
	diffuse          matrix.Vec3
	specular         matrix.Vec3
	direction        matrix.Vec3
	position         matrix.Vec3
	lastFollowPos    matrix.Vec3
	intensity        float32
	constant         float32
	linear           float32
	quadratic        float32
	cutoff           float32
	outerCutoff      float32
	lightType        LightType
	castsShadows     bool
	reset            bool
}

type LightShadowShaderData struct {
	ShaderDataBase
	LightIndex int32
}

func (t LightShadowShaderData) Size() int {
	return int(unsafe.Sizeof(ShaderDataBasic{}) - ShaderBaseDataStart)
}

func SetupLightMaterials(materialCache *MaterialCache) error {
	setupMat := func(matKey string, materialCache *MaterialCache) (*Material, error) {
		mat, err := materialCache.Material(matKey)
		if err != nil {
			slog.Error("failed to load the material", "material", matKey, "error", err)
			return nil, err
		}
		return mat, nil
	}
	if mat, err := setupMat(assets.MaterialDefinitionLightDepth, materialCache); err != nil {
		return err
	} else {
		lightDepthMaterial = weak.Make(mat)
	}
	//if lightCubeDepthMaterial, err = setupMat(assets.MaterialDefinitionLightCubeDepth, materialCache); err != nil {
	//	return err
	//}
	return nil
}

func NewLight(vr *Vulkan, assetDb *assets.Database, materialCache *MaterialCache, lightType LightType) Light {
	light := Light{
		ambient:     matrix.NewVec3(0.1, 0.1, 0.1),
		diffuse:     matrix.Vec3One(),
		specular:    matrix.Vec3One(),
		intensity:   1,
		constant:    1,
		linear:      0.0014,
		quadratic:   0.000007,
		lightType:   lightType,
		cutoff:      matrix.Cos(matrix.Deg2Rad(32.5)),
		outerCutoff: matrix.Cos(matrix.Deg2Rad(50.5)),
		reset:       true,
		renderer:    vr,
	}
	for i := range cubeMapSides {
		light.lightSpaceMatrix[i].Reset()
	}
	v30 := matrix.Vec3Zero()
	light.setupRenderPass(assetDb)
	switch light.lightType {
	case LightTypeDirectional:
		fallthrough
	default:
		light.depthMaterial = lightDepthMaterial.Value()
		light.camera = cameras.NewStandardCameraOrthographic(20, 20, 20, 20, v30)
		light.camera.SetFarPlane(lightDirectionalScaleOut * 2.0)
	case LightTypePoint:
		light.depthMaterial = lightCubeDepthMaterial.Value()
		light.camera = cameras.NewStandardCamera(lightDepthMapWidth, lightDepthMapHeight,
			lightDepthMapWidth, lightDepthMapHeight, v30)
		// Make FOV exactly large enough for each face of cubemap
		light.camera.SetFOV(90)
		light.camera.SetFarPlane(50.0)
	case LightTypeSpot:
		light.depthMaterial = lightDepthMaterial.Value()
		light.camera = cameras.NewStandardCamera(lightDepthMapWidth, lightDepthMapHeight,
			lightDepthMapWidth, lightDepthMapHeight, v30)
		light.camera.SetFOV(90)
		light.camera.SetNearPlane(0.01)
		light.camera.SetFarPlane(10.0)
	}
	return light
}

func (l *Light) ShadowMapTexture() *Texture {
	return &l.renderPass.textures[0]
}

func lightTransformDrawingToDepth(drawing *Drawing) Drawing {
	copy := *drawing
	copy.Material = lightDepthMaterial.Value()
	copy.CastsShadows = false // Shadows don't cast shadows
	sd := &LightShadowShaderData{ShaderDataBase: NewShaderDataBase()}
	drawing.ShaderData.setShadow(sd)
	copy.ShaderData = sd
	return copy
}

func lightTransformDrawingToCubeDepth(drawing *Drawing) Drawing {
	copy := *drawing
	copy.Material = lightDepthMaterial.Value()
	copy.CastsShadows = false // Shadows don't cast shadows
	sd := &LightShadowShaderData{ShaderDataBase: NewShaderDataBase()}
	copy.ShaderData = sd
	return copy
}

func (l *Light) recalculate(followCam cameras.Camera) {
	if !l.reset {
		return
	}
	if l.lightType == LightTypeDirectional {
		l.position = l.direction.Scale(-l.camera.FarPlane() * 0.5)
		if followCam != nil {
			l.lastFollowPos = matrix.NewVec3(followCam.Position().X(), 0, followCam.Position().Z())
		}
		l.position.AddAssign(l.lastFollowPos)
	}
	lookAt := l.position.Add(l.direction)
	lookAt.AddAssign(matrix.NewVec3(0.00001, 0, 0.00001))
	l.camera.SetPositionAndLookAt(l.position, lookAt)
	switch l.lightType {
	case LightTypeDirectional, LightTypeSpot:
		l.lightSpaceMatrix[0] = matrix.Mat4Multiply(l.camera.View(), l.camera.Projection())
	case LightTypePoint:
	}
	l.reset = false
}

func (l *Light) transformToGPULight() GPULight {
	g := GPULight{
		Position:  l.position,
		Direction: l.direction,
	}
	for i := range g.Matrix {
		g.Matrix[i] = l.lightSpaceMatrix[i]
	}
	return g
}

func (l *Light) transformToGPULightInfo() GPULightInfo {
	return GPULightInfo{
		Position:    l.position,
		Intensity:   l.intensity,
		Cutoff:      l.cutoff,
		Direction:   l.direction,
		OuterCutoff: l.outerCutoff,
		Ambient:     l.ambient,
		Constant:    l.constant,
		Diffuse:     l.diffuse,
		Linear:      l.linear,
		Specular:    l.specular,
		Quadratic:   l.quadratic,
		NearPlane:   l.camera.NearPlane(),
		FarPlane:    l.camera.FarPlane(),
		Type:        int32(l.lightType),
	}
}

func (l *Light) setupRenderPass(assets *assets.Database) {
	vr := l.renderer
	rp := RenderPassData{}
	if err := unmarshallJsonFile(assets, "renderer/passes/light_depth.renderpass", &rp); err != nil {
		slog.Error("failed to load light_depth.renderpass")
		return
	}
	if pass, ok := vr.renderPassCache[rp.Name]; !ok {
		rpc := rp.Compile(vr)
		if p, ok := rpc.ConstructRenderPass(vr); ok {
			vr.renderPassCache[rp.Name] = p
			l.renderPass = p
		} else {
			slog.Error("failed to load the render pass for the light", "renderPass", rp.Name)
		}
	} else {
		l.renderPass = pass
	}
}

func (l *Light) WorldSpace(followcam cameras.Camera) matrix.Vec3 {
	l.recalculate(followcam)
	return l.position
}

func (l *Light) Direction(followcam cameras.Camera) matrix.Vec3 {
	l.recalculate(followcam)
	return l.direction
}

func (l *Light) SetPosition(position matrix.Vec3) {
	if l.lightType != LightTypeDirectional {
		l.position = position
		l.reset = true
	}
}

func (l *Light) SetDirection(dir matrix.Vec3) {
	l.direction = dir
	l.reset = true
}

func (l *Light) SetIntensity(intensity float32) {
	l.intensity = intensity
	l.reset = true
}
