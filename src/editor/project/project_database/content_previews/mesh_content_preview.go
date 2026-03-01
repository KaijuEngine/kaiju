/******************************************************************************/
/* mesh_content_preview.go                                                    */
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

package content_previews

import (
	"fmt"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/collision"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
	"log/slog"
)

func (p *ContentPreviewer) renderMesh(id string) {
	defer tracing.NewRegion("ContentPreviewer.renderMesh").End()
	km, err := readMesh(id, p.ed)
	if err != nil {
		slog.Error("failed to generate a preview for mesh", "id", id, "error", err)
		p.completeProc()
		return
	}
	host := p.ed.Host()
	adjustMeshColorAndLocation(p.cam, &km)
	mesh := rendering.NewMesh("tmp", km.Verts, km.Indexes)
	sd := shader_data_registry.Create(p.mat.Shader.ShaderDataName())
	sd.(*shader_data_registry.ShaderDataEdThumbPreviewMesh).SetCamera(
		p.cam.View(), p.cam.Projection())
	draw := rendering.Drawing{
		Material:   p.mat,
		Mesh:       mesh,
		ShaderData: sd,
	}
	host.Drawings.AddDrawing(draw)
	host.RunBeforeRender(func() {
		mesh.DelayedCreate(host.Window.GpuInstance.PrimaryDevice())
		host.RunAfterFrames(1, func() {
			p.readRenderPass(host, sd, id)
		})
	})
}

func readMesh(id string, ed EditorInterface) (kaiju_mesh.KaijuMesh, error) {
	defer tracing.NewRegion("content_previews.readMesh").End()
	cc, err := ed.Cache().Read(id)
	if err != nil {
		return kaiju_mesh.KaijuMesh{}, err
	}
	if cc.Config.Type != (content_database.Mesh{}).TypeName() {
		return kaiju_mesh.KaijuMesh{},
			fmt.Errorf("can't generate a mesh preview image for content, the provided id '%s' is not a mesh", id)
	}
	data, err := ed.ProjectFileSystem().ReadFile(cc.ContentPath())
	if err != nil {
		return kaiju_mesh.KaijuMesh{}, err
	}
	return kaiju_mesh.Deserialize(data)
}

func adjustMeshColorAndLocation(cam cameras.Camera, km *kaiju_mesh.KaijuMesh) {
	defer tracing.NewRegion("content_previews.adjustMeshColorAndLocation").End()
	offset := matrix.Vec3Zero()
	points := make([]matrix.Vec3, len(km.Verts))
	for i := range km.Verts {
		km.Verts[i].Color = matrix.ColorSlateGrey()
		points[i] = km.Verts[i].Position
		offset.AddAssign(km.Verts[i].Position)
	}
	offset = offset.Shrink(matrix.Float(len(km.Verts))).Negative()
	for i := range km.Verts {
		points[i].AddAssign(offset)
	}
	box := collision.AABBFromPoints(points)
	offset = cam.Position()
	offset.AddAssign(cam.Forward().Scale(box.Size().Length() * 1.35))
	for i := range km.Verts {
		km.Verts[i].Position = points[i].Add(offset)
	}
}
