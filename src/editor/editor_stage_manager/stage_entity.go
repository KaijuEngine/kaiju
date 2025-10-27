package editor_stage_manager

import (
	"kaiju/engine"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
)

type StageEntity struct {
	engine.Entity
	StageData StageEntityEditorData
}

func (e *StageEntity) SetMaterial(mat *rendering.Material, host *engine.Host) {
	e.StageData.ShaderData.Destroy()
	e.StageData.Description.Textures = make([]string, len(mat.Textures))
	for i := range mat.Textures {
		e.StageData.Description.Textures[i] = mat.Textures[i].Key
	}
	e.StageData.ShaderData = shader_data_registry.Create(mat.Shader.ShaderDataName())
	draw := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
	}
	host.Drawings.AddDrawing(draw)
}
