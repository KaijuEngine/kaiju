/******************************************************************************/
/* common.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_designer

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
)

type FileExtension = string

const (
	shaderSrcFolder = "content/renderer/src"
	dataInputHTML   = "editor/ui/workspace/shading_workspace_data_input.go.html"
)

func showTooltip(options map[string]string, e *document.Element) {
	id := e.Attribute("data-tooltip")
	tip, ok := options[id]
	if !ok {
		return
	}
	tipElm := e.Root().FindElementById("toolTip")
	if tipElm == nil || len(tipElm.Children) == 0 {
		return
	}
	lbl := tipElm.Children[0].UI
	if !lbl.IsType(ui.ElementTypeLabel) {
		return
	}
	lbl.ToLabel().SetText(tip)
}
