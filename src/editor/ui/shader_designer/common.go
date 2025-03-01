package shader_designer

import (
	"kaiju/markup/document"
	"kaiju/ui"
)

const (
	shaderFolder     = "content/shaders/definitions"
	renderPassFolder = "content/shaders/passes"
	pipelineFolder   = "content/shaders/pipelines"
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
