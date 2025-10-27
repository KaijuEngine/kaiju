package editor_stage_manager

import (
	"kaiju/engine"
	"kaiju/matrix"
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
	// TODO:  Match up shader data to the type of material, so if it's PBR, then
	// it should use the PBR structure.
	e.StageData.ShaderData = &rendering.ShaderDataStandard{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.ColorWhite(),
	}
	draw := rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   mat,
		Mesh:       e.StageData.Mesh,
		ShaderData: e.StageData.ShaderData,
		Transform:  &e.Transform,
	}
	host.Drawings.AddDrawing(draw)
}
