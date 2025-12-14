/******************************************************************************/
/* light_data_binding.go                                                      */
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
	Ambient     matrix.Vec3 `default:"0.1,0.1,0.1"`
	Diffuse     matrix.Vec3 `default:"1,1,1"`
	Specular    matrix.Vec3 `default:"1,1,1"`
	Intensity   float32     `default:"5"`
	Constant    float32     `default:"1"`
	Linear      float32     `default:"0.0014"`
	Quadratic   float32     `default:"0.000007"`
	Cutoff      float32     `default:"0.8433914458128857"` // matrix.Cos(matrix.Deg2Rad(32.5))
	OuterCutoff float32     `default:"0.636078220277764"`  // matrix.Cos(matrix.Deg2Rad(50.5))
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
	lm.id = host.Lighting().Lights.Add(&e.Transform, light)
}

func (c *LightModule) update(deltaTime float64) {
	if !c.entity.IsActive() {
		return
	}
	light := c.host.Lighting().Lights.FindById(c.id)
	// TODO:  Only make updates if things have changed?
	light.Transform = &c.entity.Transform
	light.Target.SetPosition(light.Transform.Position())
	light.Target.SetDirection(light.Transform.Forward())
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
