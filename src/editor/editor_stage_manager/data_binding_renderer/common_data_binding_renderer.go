/******************************************************************************/
/* common_data_binding_renderer.go                                            */
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
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"weak"
)

func commonAttached(host *engine.Host, manager *editor_stage_manager.StageManager, target *editor_stage_manager.StageEntity, iconName string) {
	mat, err := host.MaterialCache().Material(assets.MaterialDefinitionEdGizmo)
	if err != nil {
		slog.Error("failed to find the basic material", "error", err)
		return
	}
	tex, err := host.TextureCache().Texture(
		"editor/textures/icons/"+iconName, rendering.TextureFilterLinear)
	if err != nil {
		slog.Error("failed to load the gizmo icon", "icon", iconName, "error", err)
		return
	}
	mat = mat.CreateInstance([]*rendering.Texture{tex})
	mesh := rendering.NewMeshQuad(host.MeshCache())
	sd := &shader_data_registry.ShaderDataUnlit{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
		UVs:            matrix.NewVec4(0, 0, 1, 1),
	}
	host.RunOnMainThread(func() {
		tex.DelayedCreate(host.Window.Renderer)
		draw := rendering.Drawing{
			Renderer:   host.Window.Renderer,
			Material:   mat,
			Mesh:       mesh,
			ShaderData: sd,
			Transform:  &target.Transform,
			ViewCuller: &host.Cameras.Primary,
		}
		host.Drawings.AddDrawing(draw)
	})
	box := collision.AABBFromTransform(&target.Transform)
	box.Extent.ScaleAssign(0.5)
	target.StageData.Bvh = collision.NewBVH([]collision.HitObject{box}, &target.Transform, target)
	manager.AddBVH(target.StageData.Bvh, &target.Transform)
	wManager := weak.Make(manager)
	target.OnDestroy.Add(func() {
		m := wManager.Value()
		if m != nil {
			m.RemoveEntityBVH(target)
		}
		sd.Destroy()
	})
}
