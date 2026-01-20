/******************************************************************************/
/* light.go                                                                   */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/rendering/vulkan_const"
	"log/slog"
	"unsafe"
	"weak"
)

const (
	MaxLocalLights           = 20
	cubeMapSides             = 6
	lightDepthMapWidth       = 4096
	lightDepthMapHeight      = 4096
	lightWidth               = 50
	lightHeight              = 50
	lightDirectionalScaleOut = 50.0
	lightShadowmapFilter     = vulkan_const.FilterLinear
	lightDepthFormat         = vulkan_const.FormatD16Unorm
	MaxCascades              = 3
	//lightDepthFormat       = vk.FormatD32Sfloat
)

var (
	lightDepthMaterial     [MaxCascades]weak.Pointer[Material]
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

type LightsForRender struct {
	Lights     []Light
	HasChanges bool
}

type Light struct {
	renderer         *Vulkan
	texture          *Texture
	camera           cameras.Camera
	renderPass       *RenderPass
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
	frameDirty       bool
}

type LightShadowShaderData struct {
	ShaderDataBase
	LightIndex int32
}

func (t LightShadowShaderData) Size() int {
	return int(unsafe.Sizeof(LightShadowShaderData{}) - ShaderBaseDataStart)
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
	mats := []string{
		assets.MaterialDefinitionLightDepth,
		assets.MaterialDefinitionLightDepthCSM1,
		assets.MaterialDefinitionLightDepthCSM2,
	}
	for i := range mats {
		if mat, err := setupMat(mats[i], materialCache); err != nil {
			return err
		} else {
			lightDepthMaterial[i] = weak.Make(mat)
		}
	}
	//if lightCubeDepthMaterial, err = setupMat(assets.MaterialDefinitionLightCubeDepth, materialCache); err != nil {
	//	return err
	//}
	return nil
}

func NewLight(vr *Vulkan, assetDb assets.Database, materialCache *MaterialCache, lightType LightType) Light {
	light := Light{
		ambient:     matrix.NewVec3(0.1, 0.1, 0.1),
		diffuse:     matrix.Vec3One(),
		specular:    matrix.Vec3One(),
		intensity:   1,
		constant:    1,
		linear:      0.0014,
		quadratic:   0.000007,
		direction:   matrix.Vec3Down(),
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
		const w, h = lightWidth, lightHeight
		light.camera = cameras.NewStandardCameraOrthographic(w, h, w, h, v30)
		light.camera.SetFarPlane(lightDirectionalScaleOut * 2.0)
	case LightTypePoint:
		light.camera = cameras.NewStandardCamera(lightDepthMapWidth, lightDepthMapHeight,
			lightDepthMapWidth, lightDepthMapHeight, v30)
		// Make FOV exactly large enough for each face of cubemap
		light.camera.SetFOV(90)
		light.camera.SetFarPlane(50.0)
	case LightTypeSpot:
		light.camera = cameras.NewStandardCamera(lightDepthMapWidth, lightDepthMapHeight,
			lightDepthMapWidth, lightDepthMapHeight, v30)
		light.camera.SetFOV(90)
		light.camera.SetNearPlane(0.01)
		light.camera.SetFarPlane(10.0)
	}
	return light
}

func (l *Light) FrameDirty() bool { return l.reset }

func (l *Light) ShadowMapTexture() *Texture {
	return &l.renderPass.textures[0]
}

func (l *Light) Type() LightType { return l.lightType }
func (l *Light) IsValid() bool   { return l.renderer != nil }

func lightTransformDrawingToDepth(drawing *Drawing, cascades uint8) Drawing {
	copy := *drawing
	copy.Material = lightDepthMaterial[cascades].Value()
	copy.Material.IsLit = false
	copy.Material.ReceivesShadows = false
	copy.Material.CastsShadows = false
	sd := &LightShadowShaderData{ShaderDataBase: NewShaderDataBase()}
	drawing.ShaderData.addShadow(sd)
	copy.ShaderData = sd
	return copy
}

func (l *Light) recalculate(camera cameras.Camera) {
	if !l.reset && !camera.IsDirty() {
		return
	}
	if l.lightType == LightTypeDirectional {
		l.position = l.direction.Scale(-camera.FarPlane() * 0.5)
		l.position.AddAssign(l.lastFollowPos)
	}
	switch l.lightType {
	case LightTypeDirectional:
		lightView := matrix.Mat4Identity()
		lightProjection := matrix.Mat4Identity()
		camView := camera.View()
		csmProjections := camera.LightFrustumCSMProjections()
		for i := range csmProjections {
			// TODO:  This shouldn't happen all the time, when the view changes,
			// might be best to store it along side the camera frustum?
			corners := collision.FrustumExtractCorners(camView, csmProjections[i])
			center := corners.Center()
			lightView.Reset()
			lightEye := center.Add(l.direction)
			lightEye.AddAssign(matrix.NewVec3(0.00001, 0, 0.00001))
			lightView.LookAt(lightEye, center, matrix.Vec3Up())
			mm := l.minMaxFromCorners(lightView, corners)
			lightProjection.Reset()
			lightProjection.Orthographic(mm.Min.X(), mm.Max.X(),
				mm.Min.Y(), mm.Max.Y(), mm.Max.Z(), mm.Min.Z())
			l.lightSpaceMatrix[i] = matrix.Mat4Multiply(lightView, lightProjection)
		}
	case LightTypePoint:
	case LightTypeSpot:
	}
	l.reset = false
}

func (l *Light) minMaxFromCorners(view matrix.Mat4, corners collision.FrustumCorners) matrix.Vec3MinMax {
	mm := matrix.NewVec3MinMax()
	for i := range corners {
		trf := matrix.Mat4MultiplyVec4(view, corners[i])
		mm.Min.SetX(min(mm.Min.X(), trf.X()))
		mm.Max.SetX(max(mm.Max.X(), trf.X()))
		mm.Min.SetY(min(mm.Min.Y(), trf.Y()))
		mm.Max.SetY(max(mm.Max.Y(), trf.Y()))
		mm.Min.SetZ(min(mm.Min.Z(), trf.Z()))
		mm.Max.SetZ(max(mm.Max.Z(), trf.Z()))
	}
	return mm
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

func (l *Light) setupRenderPass(assets assets.Database) {
	vr := l.renderer
	rp := RenderPassData{}
	if err := unmarshallJsonFile(assets, "light_depth.renderpass", &rp); err != nil {
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
	if l.lightType == LightTypeDirectional {
		return
	}
	if l.position.Equals(position) {
		return
	}
	l.position = position
	l.setDirty()
}

func (l *Light) SetDirection(dir matrix.Vec3) {
	if l.direction.Equals(dir) {
		return
	}
	l.direction = dir
	l.setDirty()
}

func (l *Light) SetIntensity(intensity float32) {
	if matrix.Approx(l.intensity, intensity) {
		return
	}
	l.intensity = intensity
	l.setDirty()
}

func (l *Light) SetConstant(constant float32) {
	if matrix.Approx(l.constant, constant) {
		return
	}
	l.constant = constant
	l.setDirty()
}

func (l *Light) SetLinear(linear float32) {
	if matrix.Approx(l.linear, linear) {
		return
	}
	l.linear = linear
	l.setDirty()
}

func (l *Light) SetQuadratic(quadratic float32) {
	if matrix.Approx(l.quadratic, quadratic) {
		return
	}
	l.quadratic = quadratic
	l.setDirty()
}

func (l *Light) SetCutoff(cutoff float32) {
	if matrix.Approx(l.cutoff, cutoff) {
		return
	}
	l.cutoff = cutoff
	l.setDirty()
}

func (l *Light) SetOuterCutoff(outerCutoff float32) {
	if matrix.Approx(l.outerCutoff, outerCutoff) {
		return
	}
	l.outerCutoff = outerCutoff
	l.setDirty()
}

func (l *Light) SetAmbient(ambient matrix.Vec3) {
	if l.ambient.Equals(ambient) {
		return
	}
	l.ambient = ambient
	l.setDirty()
}

func (l *Light) SetDiffuse(diffuse matrix.Vec3) {
	if l.diffuse.Equals(diffuse) {
		return
	}
	l.diffuse = diffuse
	l.setDirty()
}

func (l *Light) SetSpecular(specular matrix.Vec3) {
	if l.specular.Equals(specular) {
		return
	}
	l.specular = specular
	l.setDirty()
}

func (l *Light) SetCastsShadows(castsShadows bool) {
	if l.castsShadows == castsShadows {
		return
	}
	l.castsShadows = castsShadows
	// TODO:  Create or remove shadow texture
	l.setDirty()
}

func (l *Light) ResetFrameDirty() bool {
	wasReset := l.frameDirty
	l.frameDirty = false
	return wasReset
}

func (l *Light) setDirty() {
	l.frameDirty = true
	l.reset = true
}
