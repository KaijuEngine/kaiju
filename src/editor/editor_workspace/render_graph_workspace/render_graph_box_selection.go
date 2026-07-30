/******************************************************************************/
/* render_graph_box_selection.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
)

const renderGraphBoxSelectThreshold = matrix.Float(5)

func (g *renderGraph) updateBoxSelection() {
	defer tracing.NewRegion("renderGraph.updateBoxSelection").End()
	if g == nil || !g.boxSelecting || g.host == nil || g.host.Window == nil {
		return
	}
	mouse := &g.host.Window.Mouse
	current := g.graphPositionFromView(g.screenToViewPosition(mouse.ScreenPosition()))
	if mouse.Released(hid.MouseButtonLeft) {
		g.finishBoxSelection(current)
		return
	}
	if !mouse.Held(hid.MouseButtonLeft) && !mouse.Pressed(hid.MouseButtonLeft) {
		g.cancelBoxSelection()
		return
	}
	g.updateSelectionBoxVisual(current)
}

func (g *renderGraph) finishBoxSelection(current matrix.Vec2) {
	if g == nil {
		return
	}
	if g.boxStart.Distance(current) <= renderGraphBoxSelectThreshold/g.zoomValue() {
		g.SelectNodes(nil, renderGraphSelectionReplace)
		g.cancelBoxSelection()
		return
	}
	box := matrix.Vec4Area(g.boxStart.X(), g.boxStart.Y(), current.X(), current.Y())
	mode := renderGraphBoxSelectionModeFromKeyboard(nil)
	if g.host != nil && g.host.Window != nil {
		mode = renderGraphBoxSelectionModeFromKeyboard(&g.host.Window.Keyboard)
	}
	g.SelectNodes(g.nodesTouchedByBox(box), mode)
	g.cancelBoxSelection()
}

func (g *renderGraph) cancelBoxSelection() {
	if g == nil {
		return
	}
	g.boxSelecting = false
	if g.selectionBox != nil {
		g.selectionBox.Base().Hide()
	}
}

func (g *renderGraph) updateSelectionBoxVisual(current matrix.Vec2) {
	if g == nil || g.selectionBox == nil {
		return
	}
	box := matrix.Vec4Area(g.boxStart.X(), g.boxStart.Y(), current.X(), current.Y())
	viewPosition := g.viewPosition(matrix.NewVec2(box.Left(), box.Top()))
	zoom := g.zoomValue()
	width := max(matrix.Float(0.0001), box.Right()-box.Left())
	height := max(matrix.Float(0.0001), box.Bottom()-box.Top())
	base := g.selectionBox.Base()
	base.Show()
	base.Layout().SetOffset(viewPosition.X(), viewPosition.Y())
	base.Layout().Scale(matrix.Float(width*zoom), matrix.Float(height*zoom))
	base.Clean()
}

func (g *renderGraph) nodesTouchedByBox(box matrix.Vec4) []*renderGraphNode {
	if g == nil {
		return nil
	}
	nodes := make([]*renderGraphNode, 0)
	for i := range g.nodes {
		node := g.nodes[i]
		if node == nil {
			continue
		}
		if renderGraphRectsIntersect(box, node.bounds()) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (n *renderGraphNode) bounds() matrix.Vec4 {
	if n == nil {
		return matrix.Vec4Zero()
	}
	height := matrix.Float(n.height)
	if height <= 0 {
		height = matrix.Float(renderGraphNodeBaseHeight)
	}
	return matrix.NewVec4(
		n.position.X(),
		n.position.Y(),
		n.position.X()+matrix.Float(renderGraphNodeWidth),
		n.position.Y()+height,
	)
}

func renderGraphRectsIntersect(a, b matrix.Vec4) bool {
	return a.Left() <= b.Right() &&
		a.Right() >= b.Left() &&
		a.Top() <= b.Bottom() &&
		a.Bottom() >= b.Top()
}
