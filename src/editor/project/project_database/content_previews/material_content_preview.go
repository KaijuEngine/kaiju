package content_previews

import (
	"encoding/json"
	"fmt"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
)

func (p *ContentPreviewer) renderMaterial(id string) {
	mat, err := readMaterial(id, p.ed)
	if err != nil {
		slog.Error("failed to generate a preview for material", "id", id, "error", err)
		p.completeProc()
		return
	}
	host := p.ed.Host()
	mesh := rendering.NewMeshSphere(host.MeshCache(), 1, 32, 32)
	sd := shader_data_registry.Create(mat.Shader.ShaderDataName())
	draw := rendering.Drawing{
		Material:   mat,
		Mesh:       mesh,
		ShaderData: sd,
	}
	host.Drawings.AddDrawing(draw)
	host.RunBeforeRender(func() {
		mat.Shader.DelayedCreate(host.Window.Renderer, host.AssetDatabase())
		host.RunAfterFrames(1, func() {
			p.readRenderPass(host, sd, id)
		})
	})
}

func readMaterial(id string, ed EditorInterface) (*rendering.Material, error) {
	defer tracing.NewRegion("content_previews.readMaterial").End()
	cc, err := ed.Cache().Read(id)
	if err != nil {
		return nil, err
	}
	if cc.Config.Type != (content_database.Material{}).TypeName() {
		return nil, fmt.Errorf("can't generate a material preview image for content, the provided id '%s' is not a material", id)
	}
	matStr, err := ed.ProjectFileSystem().ReadFile(cc.ContentPath())
	if err != nil {
		return nil, err
	}
	key := "preview_" + id
	var materialData rendering.MaterialData
	if err := json.Unmarshal([]byte(matStr), &materialData); err != nil {
		slog.Error("failed to read the material", "material", key, "error", err)
		return nil, err
	}
	materialData.RenderPass = "ed_thumb_preview_mesh.renderpass"
	materialData.ShaderPipeline = "ed_thumb_preview_mesh.shaderpipeline"
	host := ed.Host()
	mat, err := materialData.CompileExt(host.AssetDatabase(), host.Window.Renderer, true)
	if err != nil {
		return nil, err
	}
	mat.Id = key
	return mat, nil
}
