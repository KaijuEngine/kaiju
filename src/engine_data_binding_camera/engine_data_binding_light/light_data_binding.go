package engine_data_binding_light

import (
	"kaiju/engine"
	"kaiju/engine/lighting"
	"kaiju/matrix"
	"kaiju/rendering"
)

const BindingKey = "kaiju.LightDataBinding"

func init() {
	engine.RegisterEntityData(BindingKey, LightDataBinding{})
}

type LightDataBinding struct {
	Ambient     matrix.Vec3
	Diffuse     matrix.Vec3
	Specular    matrix.Vec3
	Intensity   float32 `default:"5"`
	Constant    float32 `default:"1"`
	Linear      float32 `default:"0.0014"`
	Quadratic   float32 `default:"0.000007"`
	Cutoff      float32 `default:"0.8433914458128857"` // matrix.Cos(matrix.Deg2Rad(32.5))
	OuterCutoff float32 `default:"0.636078220277764"`  // matrix.Cos(matrix.Deg2Rad(50.5))
	//lightType    rendering.LightType
	CastsShadows bool
}

type LightModule struct {
	id       lighting.EntryId
	entity   *engine.Entity
	host     *engine.Host
	updateId engine.UpdateId
	Data     LightDataBinding
}

func (c LightDataBinding) Init(e *engine.Entity, host *engine.Host) {
	light := rendering.NewLight(host.Window.Renderer.(*rendering.Vulkan),
		host.AssetDatabase(), host.MaterialCache(), rendering.LightTypeDirectional)
	light.SetDirection(matrix.Vec3Forward())
	light.SetPosition(matrix.Vec3Zero())
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
	lm.id = host.Lighting().Lights.Add(e.Transform.Position(), light)
}

func (c *LightModule) update(deltaTime float64) {
	if !c.entity.IsActive() {
		return
	}
	t := &c.entity.Transform
	light := c.host.Lighting().Lights.FindById(c.id)
	// TODO:  Only make updates if things have changed?
	light.Position = t.Position()
	light.Target.SetPosition(light.Position)
	light.Target.SetDirection(t.Forward())
	light.Target.SetDirection(matrix.Vec3Forward())
	light.Target.SetPosition(matrix.Vec3Zero())
	light.Target.SetAmbient(c.Data.Ambient)
	light.Target.SetDiffuse(c.Data.Diffuse)
	light.Target.SetSpecular(c.Data.Specular)
	light.Target.SetIntensity(c.Data.Intensity)
	light.Target.SetConstant(c.Data.Constant)
	light.Target.SetLinear(c.Data.Linear)
	light.Target.SetQuadratic(c.Data.Quadratic)
	light.Target.SetCutoff(c.Data.Cutoff)
	light.Target.SetOuterCutoff(c.Data.OuterCutoff)
	light.Target.SetCastsShadows(c.Data.CastsShadows)
}
