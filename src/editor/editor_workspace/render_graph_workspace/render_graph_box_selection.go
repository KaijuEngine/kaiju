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

const shaderGraphBoxSelectThreshold = matrix.Float(5)

func (g *shaderGraph) updateBoxSelection() {
	defer tracing.NewRegion("shaderGraph.updateBoxSelection").End()
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

func (g *shaderGraph) finishBoxSelection(current matrix.Vec2) {
	if g == nil {
		return
	}
	if g.boxStart.Distance(current) <= shaderGraphBoxSelectThreshold/g.zoomValue() {
		g.SelectNodes(nil, shaderGraphSelectionReplace)
		g.cancelBoxSelection()
		return
	}
	box := matrix.Vec4Area(g.boxStart.X(), g.boxStart.Y(), current.X(), current.Y())
	mode := shaderGraphBoxSelectionModeFromKeyboard(nil)
	if g.host != nil && g.host.Window != nil {
		mode = shaderGraphBoxSelectionModeFromKeyboard(&g.host.Window.Keyboard)
	}
	g.SelectNodes(g.nodesTouchedByBox(box), mode)
	g.cancelBoxSelection()
}

func (g *shaderGraph) cancelBoxSelection() {
	if g == nil {
		return
	}
	g.boxSelecting = false
	if g.selectionBox != nil {
		g.selectionBox.Base().Hide()
	}
}

func (g *shaderGraph) updateSelectionBoxVisual(current matrix.Vec2) {
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
	base.Layout().Scale(float32(width*zoom), float32(height*zoom))
	base.Clean()
}

func (g *shaderGraph) nodesTouchedByBox(box matrix.Vec4) []*shaderGraphNode {
	if g == nil {
		return nil
	}
	nodes := make([]*shaderGraphNode, 0)
	for i := range g.nodes {
		node := g.nodes[i]
		if node == nil {
			continue
		}
		if shaderGraphRectsIntersect(box, node.bounds()) {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (n *shaderGraphNode) bounds() matrix.Vec4 {
	if n == nil {
		return matrix.Vec4Zero()
	}
	height := matrix.Float(n.height)
	if height <= 0 {
		height = matrix.Float(shaderGraphNodeBaseHeight)
	}
	return matrix.NewVec4(
		n.position.X(),
		n.position.Y(),
		n.position.X()+matrix.Float(shaderGraphNodeWidth),
		n.position.Y()+height,
	)
}

func shaderGraphRectsIntersect(a, b matrix.Vec4) bool {
	return a.Left() <= b.Right() &&
		a.Right() >= b.Left() &&
		a.Top() <= b.Bottom() &&
		a.Bottom() >= b.Top()
}
