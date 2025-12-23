/******************************************************************************/
/* light_data_binding_renderer.go                                             */
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

package data_binding_renderer

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/lighting"
	"kaiju/engine_entity_data/engine_entity_data_light"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
)

func init() {
	AddRenderer(engine_entity_data_light.BindingKey, &LightEntityDataRenderer{
		Lights: make(map[*editor_stage_manager.StageEntity]lightEntityDataDrawing),
	})
}

type LightEntityDataRenderer struct {
	Lights map[*editor_stage_manager.StageEntity]lightEntityDataDrawing
}

type lightEntityDataDrawing struct {
	icon  rendering.DrawInstance
	lines rendering.DrawInstance
	light *lighting.LightEntry
}

func (c *LightEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("LightEntityDataRenderer.Attached").End()
	icon := commonAttached(host, manager, target, "light.png")
	if _, ok := c.Lights[target]; ok {
		return
	}
	lightType := rendering.LightType(data.FieldValueByName("Type").(int))
	l := rendering.NewLight(host.Window.Renderer.(*rendering.Vulkan),
		host.AssetDatabase(), host.MaterialCache(), lightType)
	l.SetPosition(target.Transform.WorldPosition())
	l.SetDirection(target.Transform.Up().Negative())
	l.SetAmbient(data.FieldValueByName("Ambient").(matrix.Vec3))
	l.SetDiffuse(data.FieldValueByName("Diffuse").(matrix.Vec3))
	l.SetSpecular(data.FieldValueByName("Specular").(matrix.Vec3))
	l.SetIntensity(float32(data.FieldValueByName("Intensity").(float32)))
	l.SetConstant(float32(data.FieldValueByName("Constant").(float32)))
	l.SetLinear(float32(data.FieldValueByName("Linear").(float32)))
	l.SetQuadratic(float32(data.FieldValueByName("Quadratic").(float32)))
	l.SetCutoff(float32(data.FieldValueByName("Cutoff").(float32)))
	l.SetOuterCutoff(float32(data.FieldValueByName("OuterCutoff").(float32)))
	l.SetCastsShadows(data.FieldValueByName("CastsShadows").(bool))
	c.Lights[target] = lightEntityDataDrawing{
		icon:  icon,
		lines: c.createLines(host, &target.Transform),
		light: host.Lighting().Lights.Add(&target.Transform, l),
	}
	target.OnActivate.Add(func() {
		if d, ok := c.Lights[target]; ok {
			d.icon.Activate()
			d.light = host.Lighting().Lights.Add(&target.Transform, l)
			c.Lights[target] = d
		}
	})
	target.OnDeactivate.Add(func() {
		if d, ok := c.Lights[target]; ok {
			d.icon.Deactivate()
			host.Lighting().Lights.Remove(d.light)
			d.light = nil
			c.Lights[target] = d
		}
	})
}

func (c *LightEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("LightEntityDataRenderer.Detatched").End()
	if d, ok := c.Lights[target]; ok {
		if d.light != nil {
			host.Lighting().Lights.Remove(d.light)
		}
		delete(c.Lights, target)
	}
}

func (c *LightEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("LightEntityDataRenderer.Show").End()
	if d, ok := c.Lights[target]; ok {
		d.lines.Activate()
	}
}

func (c *LightEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("LightEntityDataRenderer.Hide").End()
	if d, ok := c.Lights[target]; ok {
		d.lines.Deactivate()
	}
}

func (c *LightEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	l := c.Lights[target]
	lightType := rendering.LightType(data.FieldValueByName("Type").(int))
	if l.light.Type() != lightType {
		l.light.Light = rendering.NewLight(host.Window.Renderer.(*rendering.Vulkan),
			host.AssetDatabase(), host.MaterialCache(), lightType)
	}
	l.light.Light.SetPosition(l.light.Transform.WorldPosition())
	l.light.Light.SetDirection(l.light.Transform.Up().Negative())
	l.light.Light.SetAmbient(data.FieldValueByName("Ambient").(matrix.Vec3))
	l.light.Light.SetDiffuse(data.FieldValueByName("Diffuse").(matrix.Vec3))
	l.light.Light.SetSpecular(data.FieldValueByName("Specular").(matrix.Vec3))
	l.light.Light.SetIntensity(float32(data.FieldValueByName("Intensity").(float32)))
	l.light.Light.SetConstant(float32(data.FieldValueByName("Constant").(float32)))
	l.light.Light.SetLinear(float32(data.FieldValueByName("Linear").(float32)))
	l.light.Light.SetQuadratic(float32(data.FieldValueByName("Quadratic").(float32)))
	l.light.Light.SetCutoff(float32(data.FieldValueByName("Cutoff").(float32)))
	l.light.Light.SetOuterCutoff(float32(data.FieldValueByName("OuterCutoff").(float32)))
	l.light.Light.SetCastsShadows(data.FieldValueByName("CastsShadows").(bool))
}

func (c *LightEntityDataRenderer) createLines(host *engine.Host, transform *matrix.Transform) rendering.DrawInstance {
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load the grid material", "error", err)
		return nil
	}
	points := []matrix.Vec3{
		matrix.NewVec3(0, 0, 0), // Center
		matrix.NewVec3(0, -1.5, 0),
		matrix.NewVec3(-0.2, 0, -0.2),
		matrix.NewVec3(-0.2, -1, -0.2),
		matrix.NewVec3(-0.2, 0, 0.2),
		matrix.NewVec3(-0.2, -1, 0.2),
		matrix.NewVec3(0.2, 0, -0.2),
		matrix.NewVec3(0.2, -1, -0.2),
		matrix.NewVec3(0.2, 0, 0.2),
		matrix.NewVec3(0.2, -1, 0.2),
	}
	const key = "ed_directional_light_lines"
	var grid *rendering.Mesh
	var ok bool
	if grid, ok = host.MeshCache().FindMesh(key); !ok {
		grid = rendering.NewMeshGrid(host.MeshCache(), key,
			points, matrix.Color{1, 1, 1, 1})
	}
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	gsd := sd.(*shader_data_registry.ShaderDataEdTransformWire)
	gsd.Color = matrix.NewColor(1, 1, 1, 1)
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       grid,
		ShaderData: gsd,
		Transform:  transform,
		ViewCuller: &host.Cameras.Primary,
	})
	return sd
}
