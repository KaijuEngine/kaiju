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

package data_binding_renderer

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/cameras"
	"kaiju/engine_data_bindings"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"

	"github.com/KaijuEngine/uuid"
)

func init() {
	AddRenderer(engine_data_bindings.CameraDataBindingKey, &CameraDataBindingRenderer{
		Frustums: make(map[*editor_stage_manager.StageEntity]cameraDataBindingDrawing),
	})
}

type CameraDataBindingRenderer struct {
	Frustums map[*editor_stage_manager.StageEntity]cameraDataBindingDrawing
}

type cameraDataBindingDrawing struct {
	key string
	sd  rendering.DrawInstance
}

func (c *CameraDataBindingRenderer) Show(host *engine.Host, target *editor_stage_manager.StageEntity, data *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraDataBindingRenderer.Show").End()
	if _, ok := c.Frustums[target]; ok {
		slog.Error("there is an internal error in state for the editor's CameraDataBindingRenderer, show was called before any hide happened. Double selected the same target?")
		c.Hide(host, target, data)
	}
	w, h := float32(host.Window.Width()), float32(host.Window.Height())
	cam := cameras.NewStandardCamera(w, h, w, h, target.Transform.Position())
	cam.SetProperties(
		data.FieldValueByName("FOV").(float32),
		data.FieldValueByName("NearPlane").(float32),
		data.FieldValueByName("FarPlane").(float32),
		w, h,
	)
	frustum := rendering.NewMeshFrustum(host.MeshCache(), uuid.NewString(), cam.InverseProjection())
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionEdTransformWire)
	if err != nil {
		slog.Error("failed to load transform wire material", "error", err)
		return
	}
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	sd.(*shader_data_registry.ShaderDataEdTransformWire).Color = matrix.ColorWhite()
	host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   material,
		Mesh:       frustum,
		ShaderData: sd,
		Transform:  &target.Transform,
	})
	c.Frustums[target] = cameraDataBindingDrawing{frustum.Key(), sd}
}

func (c *CameraDataBindingRenderer) Hide(host *engine.Host, target *editor_stage_manager.StageEntity, _ *entity_data_binding.EntityDataEntry) {
	defer tracing.NewRegion("CameraDataBindingRenderer.Hide").End()
	if d, ok := c.Frustums[target]; ok {
		d.sd.Destroy()
		host.MeshCache().RemoveMesh(d.key)
		delete(c.Frustums, target)
	}
}
