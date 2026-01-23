package content_previews

import (
	"fmt"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/engine/cameras"
	"kaiju/engine/collision"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"kaiju/rendering/loaders/kaiju_mesh"
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
		mesh.DelayedCreate(host.Window.Renderer)
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
