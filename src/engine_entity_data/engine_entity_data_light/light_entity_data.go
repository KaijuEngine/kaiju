/******************************************************************************/
/* light_data_binding.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine_entity_data_light

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/engine/lighting"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

var bindingKey = ""

type LightType int

const (
	LightTypeDirectional = LightType(iota)
	LightTypePoint
	LightTypeSpot
)

func init() {
	engine.RegisterEntityData(LightEntityData{})
}

func BindingKey() string {
	if bindingKey == "" {
		bindingKey = pod.QualifiedNameForLayout(LightEntityData{})
	}
	return bindingKey
}

type LightEntityData struct {
	Ambient      matrix.Vec3 `default:"0.1,0.1,0.1"`
	Diffuse      matrix.Vec3 `default:"1,1,1"`
	Specular     matrix.Vec3 `default:"1,1,1"`
	Intensity    float32     `default:"5"`
	Constant     float32     `default:"1"`
	Linear       float32     `default:"0.0014"`
	Quadratic    float32     `default:"0.000007"`
	Cutoff       float32     `default:"0.8433914458128857"` // matrix.Cos(matrix.Deg2Rad(32.5))
	OuterCutoff  float32     `default:"0.636078220277764"`  // matrix.Cos(matrix.Deg2Rad(50.5))
	Type         LightType
	CastsShadows bool
}

// WithLegacyColorDefaults repairs light bindings written by editor versions
// that failed to apply matrix vector default tags. That defect serialized all
// three color fields as zero, making the light permanently black after reload.
// A partial zero value is left alone so authored color choices are preserved.
func (c LightEntityData) WithLegacyColorDefaults() LightEntityData {
	zero := matrix.Vec3Zero()
	if c.Ambient == zero && c.Diffuse == zero && c.Specular == zero {
		c.Ambient = matrix.NewVec3(0.1, 0.1, 0.1)
		c.Diffuse = matrix.Vec3One()
		c.Specular = matrix.Vec3One()
	}
	return c
}

type LightModule struct {
	lightEntry *lighting.LightEntry
	entity     *engine.Entity
	host       *engine.Host
	updateId   engine.UpdateId
	Data       LightEntityData
}

func (c LightEntityData) Init(e *engine.Entity, host *engine.Host) {
	c = c.WithLegacyColorDefaults()
	light := rendering.NewLight(host.Window.GpuInstance.PrimaryDevice(),
		host.AssetDatabase(), host.MaterialCache(),
		rendering.LightType(c.Type))
	light.SetPosition(e.Transform.WorldPosition())
	light.SetDirection(e.Transform.Up().Negative())
	light.SetAmbient(c.Ambient)
	light.SetDiffuse(c.Diffuse)
	light.SetSpecular(c.Specular)
	light.SetIntensity(c.Intensity)
	light.SetConstant(c.Constant)
	light.SetLinear(c.Linear)
	light.SetQuadratic(c.Quadratic)
	light.SetCutoff(c.Cutoff)
	light.SetOuterCutoff(c.OuterCutoff)
	light.SetCastsShadows(c.CastsShadows)
	lm := &LightModule{
		entity: e,
		host:   host,
		Data:   c,
	}
	e.AddNamedData("LightModule", lm)
	lm.updateId = host.Updater.AddUpdate(lm.update)
	lm.lightEntry = host.Lighting().Lights.Add(&e.Transform, light)
}

func (c *LightModule) update(deltaTime float64) {
	if !c.entity.IsActive() {
		return
	}
	light := c.lightEntry
	// TODO:  Only make updates if things have changed?
	light.Transform = &c.entity.Transform
	light.SetPosition(light.Transform.WorldPosition())
	light.SetDirection(light.Transform.Up().Negative())
	light.Light.SetAmbient(c.Data.Ambient)
	light.Light.SetDiffuse(c.Data.Diffuse)
	light.Light.SetSpecular(c.Data.Specular)
	light.Light.SetIntensity(c.Data.Intensity)
	light.Light.SetConstant(c.Data.Constant)
	light.Light.SetLinear(c.Data.Linear)
	light.Light.SetQuadratic(c.Data.Quadratic)
	light.Light.SetCutoff(c.Data.Cutoff)
	light.Light.SetOuterCutoff(c.Data.OuterCutoff)
	light.Light.SetCastsShadows(c.Data.CastsShadows)
}
