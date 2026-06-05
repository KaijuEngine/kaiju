/******************************************************************************/
/* shading_workspace_layout.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
)

func (w *ShadingWorkspace) applyLayout() {
	if w.Host == nil || w.Host.Window == nil {
		return
	}
	windowWidth := float32(w.Host.Window.Width())
	windowHeight := float32(w.Host.Window.Height())
	contentHeight := max(1, windowHeight-shadingGraphMenuBarHeight-shadingGraphStatusBarHeight)
	stageHeight := max(1, contentHeight*0.5)
	graphHeight := max(1, contentHeight-stageHeight)
	graphTop := shadingGraphMenuBarHeight + stageHeight

	w.applyPanelLayout(w.stageViewport, 0, shadingGraphMenuBarHeight, windowWidth, stageHeight, 2)
	w.applyPanelLayout(w.shaderGraphArea, 0, graphTop, windowWidth, graphHeight, 1)
}

func (w *ShadingWorkspace) applyPanelLayout(element *document.Element, x, y, width, height, z float32) {
	if element == nil || element.UI == nil {
		return
	}
	layout := element.UI.Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.Scale(width, height)
	layout.SetOffset(x, y)
	layout.SetZ(z)
}
