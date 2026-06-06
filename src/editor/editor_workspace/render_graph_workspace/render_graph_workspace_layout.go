/******************************************************************************/
/* render_graph_workspace_layout.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
)

const renderGraphSidePanelWidth = float32(260.0)

func (w *RenderGraphWorkspace) applyLayout() {
	if w.Host == nil || w.Host.Window == nil {
		return
	}
	windowWidth := float32(w.Host.Window.Width())
	windowHeight := float32(w.Host.Window.Height())
	viewX := renderGraphSidePanelWidth
	viewWidth := max(1, windowWidth-viewX)
	contentHeight := max(1, windowHeight-renderGraphMenuBarHeight-renderGraphStatusBarHeight)
	stageHeight := max(1, contentHeight*0.5)
	graphHeight := max(1, contentHeight-stageHeight)
	graphTop := renderGraphMenuBarHeight + stageHeight

	w.applyPanelLayout(w.sidePanel, 0, 0, renderGraphSidePanelWidth, windowHeight, 5)
	w.applyStageViewportLayout(w.stageViewport, viewX, renderGraphMenuBarHeight, viewWidth, stageHeight, 2)
	w.applyPanelLayout(w.renderGraphArea, viewX, graphTop, viewWidth, graphHeight, 1)
	w.graph.SetViewport(viewX, graphTop, viewWidth, graphHeight)
}

func (w *RenderGraphWorkspace) applyPanelLayout(element *document.Element, x, y, width, height, z float32) {
	if element == nil || element.UI == nil {
		return
	}
	layout := element.UI.Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.Scale(width, height)
	layout.SetOffset(x, y)
	layout.SetZ(z)
}

func (w *RenderGraphWorkspace) applyStageViewportLayout(element *document.Element, left, top, width, height, z float32) {
	if element == nil || element.UI == nil || w.Host == nil || w.Host.Window == nil {
		return
	}
	windowWidth := float32(w.Host.Window.Width())
	x := left - windowWidth*0.5 + width*0.5
	y := top
	layout := element.UI.Layout()
	layout.SetPositioning(ui.PositioningAbsolute)
	layout.Scale(width, height)
	layout.SetOffset(x, y)
	layout.SetZ(z)
}
