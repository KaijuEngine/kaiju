/******************************************************************************/
/* camera_data_binding_renderer.go                                            */
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
	"kaiju/engine/cameras"
	"kaiju/engine_entity_data/engine_entity_data_camera"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
)

func init() {
	AddRenderer(engine_entity_data_camera.BindingKey, &CameraEntityDataRenderer{
		Frustums: make(map[*editor_stage_manager.StageEntity]cameraDataBindingDrawing),
	})
}

type CameraEntityDataRenderer struct {
	Frustums map[*editor_stage_manager.StageEntity]cameraDataBindingDrawing
}

type cameraDataBindingDrawing struct {
	key  string
	sd   rendering.DrawInstance
	icon rendering.DrawInstance
}

func (c *CameraEntityDataRenderer) Attached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Attached").End()
	icon := commonAttached(host, manager, target, "camera.png")
	if _, ok := c.Frustums[target]; ok {
		slog.Error("there is an internal error in state for the editor's CameraEntityDataRenderer, show was called before any hide happened. Double selected the same target?")
		c.Detatched(host, manager, target, data)
	}
	w, h := float32(host.Window.Width()), float32(host.Window.Height())
	cam := cameras.NewStandardCamera(w, h, w, h, target.Transform.Position())
	cam.SetProperties(
		data.FieldValueByName("FOV").(float32),
		data.FieldValueByName("NearPlane").(float32),
		data.FieldValueByName("FarPlane").(float32),
		w, h,
	)
	frustum := rendering.NewMeshFrustumBox(host.MeshCache(), cam.InverseProjection())
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdFrustumWire)
	if err != nil {
		slog.Error("failed to load transform wire material", "error", err)
		return
	}
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	sd.(*shader_data_registry.ShaderDataEdFrustumWire).Color = matrix.ColorWhite()
	sd.(*shader_data_registry.ShaderDataEdFrustumWire).FrustumProjection = cam.InverseProjection()
	host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       frustum,
		ShaderData: sd,
		Transform:  &target.Transform,
		ViewCuller: &host.Cameras.Primary,
	})
	c.Frustums[target] = cameraDataBindingDrawing{frustum.Key(), sd, icon}
	target.OnActivate.Add(func() {
		if d, ok := c.Frustums[target]; ok {
			d.icon.Activate()
			d.sd.Activate()
		}
	})
	target.OnDeactivate.Add(func() {
		if d, ok := c.Frustums[target]; ok {
			d.icon.Deactivate()
			d.sd.Deactivate()
		}
	})
}

func (c *CameraEntityDataRenderer) Detatched(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Detatched").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Destroy()
		d.icon.Destroy()
		host.MeshCache().RemoveMesh(d.key)
		delete(c.Frustums, target)
	}
}

func (c *CameraEntityDataRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Show").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Activate()
	}
}

func (c *CameraEntityDataRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraEntityDataRenderer.Hide").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Deactivate()
	}
}

func (c *CameraEntityDataRenderer) Update(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	if t, ok := c.Frustums[target]; ok {
		w := float32(data.FieldValueByName("Width").(float32))
		h := float32(data.FieldValueByName("Height").(float32))
		if w <= 0 {
			w = float32(host.Window.Width())
		}
		if h <= 0 {
			h = float32(host.Window.Height())
		}
		var cam cameras.Camera
		camType := engine_entity_data_camera.CameraType(data.FieldValueByName("Type").(int))
		switch camType {
		case engine_entity_data_camera.CameraTypeOrthographic:
			cam = cameras.NewStandardCameraOrthographic(w, h, w, h, target.Transform.Position())
		case engine_entity_data_camera.CameraTypeTurntable:
			cam = cameras.ToTurntable(cameras.NewStandardCamera(w, h, w, h, target.Transform.Position()))
		case engine_entity_data_camera.CameraTypePerspective:
			fallthrough
		default:
			cam = cameras.NewStandardCamera(w, h, w, h, target.Transform.Position())
		}
		cam.SetProperties(
			data.FieldValueByName("FOV").(float32),
			data.FieldValueByName("NearPlane").(float32),
			data.FieldValueByName("FarPlane").(float32),
			w, h,
		)
		t.sd.(*shader_data_registry.ShaderDataEdFrustumWire).FrustumProjection = cam.InverseProjection()
	}
}
