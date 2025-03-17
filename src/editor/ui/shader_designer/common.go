package shader_designer

import (
	"kaiju/engine/ui/markup/document"
	"kaiju/engine/ui"
)

const (
	shaderSrcFolder  = "content/renderer/src"
	shaderSpvFolder  = "content/renderer/spv"
	shaderFolder     = "content/renderer/shaders"
	renderPassFolder = "content/renderer/passes"
	pipelineFolder   = "content/renderer/pipelines"
	materialFolder   = "content/renderer/materials"
	texturesFolder   = "content/textures"
)

func showTooltip(options map[string]string, e *document.Element) {
	id := e.Attribute("data-tooltip")
	tip, ok := options[id]
	if !ok {
		return
	}
	tipElm := e.Root().FindElementById("ToolTip")
	if tipElm == nil || len(tipElm.Children) == 0 {
		return
	}
	lbl := tipElm.Children[0].UI
	if !lbl.IsType(ui.ElementTypeLabel) {
		return
	}
	lbl.ToLabel().SetText(tip)
}
