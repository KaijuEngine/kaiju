/******************************************************************************/
/* mesh_content_preview.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_previews

import (
	"fmt"
	"log/slog"

	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/engine/cameras"
	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
	"kaijuengine.com/rendering/loaders/kaiju_mesh"
)

func (p *ContentPreviewer) renderMesh(id string) {
	defer tracing.NewRegion("ContentPreviewer.renderMesh").End()
	set, err := readMeshSet(id, p.ed)
	if err != nil {
		slog.Error("failed to generate a preview for mesh", "id", id, "error", err)
		p.completeProc()
		return
	}
	host := p.ed.Host()
	adjustMeshSetColorAndLocation(p.cam, &set)
	meshes := make([]*rendering.Mesh, 0, len(set.Meshes))
	shaderData := make([]rendering.DrawInstance, 0, len(set.Meshes))
	for i := range set.Meshes {
		km := &set.Meshes[i]
		if len(km.Verts) == 0 || len(km.Indexes) == 0 {
			continue
		}
		mesh := rendering.NewMesh(fmt.Sprintf("tmp_%s_%d", id, i), km.Verts, km.Indexes)
		sd := shader_data_registry.Create(p.mat.Shader.DrawInstanceDataName())
		sd.(*shader_data_registry.ShaderDataEdThumbPreviewMesh).SetCamera(
			p.cam.View(), p.cam.Projection())
		draw := rendering.Drawing{
			Material:   p.mat,
			Mesh:       mesh,
			ShaderData: sd,
		}
		host.Drawings.AddDrawing(draw)
		meshes = append(meshes, mesh)
		shaderData = append(shaderData, sd)
	}
	if len(meshes) == 0 {
		slog.Error("failed to generate a preview for mesh with no drawable submeshes", "id", id)
		p.completeProc()
		return
	}
	host.RunBeforeRender(func() {
		for _, mesh := range meshes {
			mesh.DelayedCreate(host.Window.GpuInstance.PrimaryDevice())
		}
		p.readRenderPassAfterNextRender(host, id, shaderData...)
	})
}

func readMeshSet(id string, ed EditorInterface) (kaiju_mesh.KaijuMeshSet, error) {
	defer tracing.NewRegion("content_previews.readMeshSet").End()
	ref := kaiju_mesh.ParseMeshRef(id)
	cc, err := ed.Cache().Read(ref.Asset)
	if err != nil {
		return kaiju_mesh.KaijuMeshSet{}, err
	}
	if cc.Config.Type != (content_database.Mesh{}).TypeName() {
		return kaiju_mesh.KaijuMeshSet{},
			fmt.Errorf("can't generate a mesh preview image for content, the provided id '%s' is not a mesh", id)
	}
	data, err := ed.ProjectFileSystem().ReadFile(cc.ContentPath())
	if err != nil {
		return kaiju_mesh.KaijuMeshSet{}, err
	}
	set, err := kaiju_mesh.DeserializeSet(data)
	if err != nil {
		return kaiju_mesh.KaijuMeshSet{}, err
	}
	if ref.Key == "" {
		return set, nil
	}
	mesh, ok := set.MeshByKey(ref.Key)
	if !ok {
		return kaiju_mesh.KaijuMeshSet{}, fmt.Errorf("mesh %q not found in %q", ref.Key, ref.Asset)
	}
	return kaiju_mesh.KaijuMeshSet{Name: set.Name, Meshes: []kaiju_mesh.KaijuMesh{mesh}}, nil
}

func adjustMeshColorAndLocation(cam cameras.Camera, km *kaiju_mesh.KaijuMesh) {
	defer tracing.NewRegion("content_previews.adjustMeshColorAndLocation").End()
	set := kaiju_mesh.KaijuMeshSet{Meshes: []kaiju_mesh.KaijuMesh{*km}}
	adjustMeshSetColorAndLocation(cam, &set)
	if len(set.Meshes) > 0 {
		*km = set.Meshes[0]
	}
}

func adjustMeshSetColorAndLocation(cam cameras.Camera, set *kaiju_mesh.KaijuMeshSet) {
	defer tracing.NewRegion("content_previews.adjustMeshSetColorAndLocation").End()
	pointCount := 0
	for meshIndex := range set.Meshes {
		pointCount += len(set.Meshes[meshIndex].Verts)
	}
	if pointCount == 0 {
		return
	}
	points := make([]matrix.Vec3, 0, pointCount)
	center := matrix.Vec3Zero()
	for meshIndex := range set.Meshes {
		km := &set.Meshes[meshIndex]
		for vertIndex := range km.Verts {
			km.Verts[vertIndex].Color = matrix.ColorSlateGrey()
			points = append(points, km.Verts[vertIndex].Position)
			center.AddAssign(km.Verts[vertIndex].Position)
		}
	}
	center = center.Shrink(matrix.Float(len(points)))
	centerOffset := center.Negative()
	for i := range points {
		points[i].AddAssign(centerOffset)
	}
	box := graviton.AABBFromPoints(points)
	previewOffset := cam.Position()
	previewOffset.AddAssign(cam.Forward().Scale(box.Size().Length() * 1.35))
	for meshIndex := range set.Meshes {
		km := &set.Meshes[meshIndex]
		for vertIndex := range km.Verts {
			km.Verts[vertIndex].Position.AddAssign(centerOffset)
			km.Verts[vertIndex].Position.AddAssign(previewOffset)
		}
	}
}
