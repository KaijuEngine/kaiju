/******************************************************************************/
/* shader_graph_view.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

func (g *shaderGraph) updatePan() {
	if g.host == nil || g.host.Window == nil || g.root == nil {
		return
	}
	mouse := &g.host.Window.Mouse
	current := mouse.ScreenPosition()
	if !g.hasPanMouse {
		g.panMouse = current
		g.hasPanMouse = true
		return
	}
	delta := current.Subtract(g.panMouse)
	g.panMouse = current

	if g.uiMan.Group.IsFocusedOnInput() {
		g.panning = false
		return
	}

	middleHeld := mouse.Pressed(hid.MouseButtonMiddle) ||
		mouse.Held(hid.MouseButtonMiddle)
	spaceHeld := g.host.Window.Keyboard.KeyHeld(hid.KeyboardKeySpace)
	wantsPan := middleHeld || spaceHeld
	if !wantsPan {
		g.panning = false
		return
	}
	if !g.panning && !g.screenPositionInside(current) {
		return
	}
	g.panning = true
	g.stopNodeDrags()
	if delta.IsZero() {
		return
	}
	if middleHeld || (spaceHeld && mouse.Moved()) {
		g.pan = g.pan.Add(delta)
		g.applyViewOffsets()
	}
}

func (g *shaderGraph) updateZoom() {
	if g == nil || g.host == nil || g.host.Window == nil || g.root == nil {
		return
	}
	if g.uiMan.Group.IsFocusedOnInput() {
		return
	}
	mouse := &g.host.Window.Mouse
	mousePosition := mouse.ScreenPosition()
	if !mouse.Scrolled() || !g.screenPositionInside(mousePosition) {
		return
	}
	if g.zoomBlocked != nil && g.zoomBlocked(mousePosition) {
		return
	}
	scroll := matrix.Float(mouse.ScrollY)
	if matrix.Approx(scroll, 0) {
		scroll = matrix.Float(mouse.ScrollX)
	}
	if matrix.Approx(scroll, 0) {
		return
	}
	factor := matrix.Float(1) + matrix.Abs(scroll)*shaderGraphZoomStep
	next := g.zoomValue()
	if scroll > 0 {
		next *= factor
	} else {
		next /= factor
	}
	g.setZoomAroundViewPosition(next, g.screenToViewPosition(mousePosition))
}

func (g *shaderGraph) isPanInputHeld() bool {
	if g == nil || g.host == nil || g.host.Window == nil {
		return false
	}
	mouse := &g.host.Window.Mouse
	keyboard := &g.host.Window.Keyboard
	return mouse.Pressed(hid.MouseButtonMiddle) ||
		mouse.Held(hid.MouseButtonMiddle) ||
		keyboard.KeyHeld(hid.KeyboardKeySpace)
}

func (g *shaderGraph) isAltInputHeld() bool {
	return g != nil && g.host != nil && g.host.Window != nil &&
		g.host.Window.Keyboard.HasAlt()
}

func (g *shaderGraph) stopNodeDrags() {
	for i := range g.nodes {
		g.nodes[i].stopDrag()
	}
	for i := range g.comments {
		g.comments[i].stopInteraction()
	}
}

func (g *shaderGraph) applyViewOffsets() {
	for i := range g.comments {
		g.comments[i].applyViewOffset()
	}
	for i := range g.nodes {
		g.nodes[i].applyViewOffset()
	}
}

func (g *shaderGraph) CenterView() {
	if g == nil {
		return
	}
	g.pan = matrix.Vec2Zero()
	g.zoom = 1
	g.applyViewOffsets()
}

func (g *shaderGraph) FocusSelection() bool {
	if g == nil || g.root == nil {
		return false
	}
	bounds, ok := shaderGraphNodesBounds(g.selected)
	if !ok && g.selectedComment != nil {
		bounds = g.selectedComment.bounds()
		ok = true
	}
	if !ok {
		return false
	}
	g.focusBounds(bounds, g.root.Base().Layout().PixelSize())
	return true
}

func (g *shaderGraph) focusBounds(bounds matrix.Vec4, viewportSize matrix.Vec2) {
	if g == nil {
		return
	}
	center := matrix.NewVec2(
		(bounds.Left()+bounds.Right())*0.5,
		(bounds.Top()+bounds.Bottom())*0.5,
	)
	viewportCenter := viewportSize.Scale(0.5)
	g.pan = viewportCenter.Subtract(center.Scale(g.zoomValue()))
	g.applyViewOffsets()
}

func shaderGraphNodesBounds(nodes []*shaderGraphNode) (matrix.Vec4, bool) {
	var bounds matrix.Vec4
	hasBounds := false
	for i := range nodes {
		node := nodes[i]
		if node == nil {
			continue
		}
		nodeBounds := node.bounds()
		if !hasBounds {
			bounds = nodeBounds
			hasBounds = true
			continue
		}
		bounds.SetX(matrix.Min(bounds.Left(), nodeBounds.Left()))
		bounds.SetY(matrix.Min(bounds.Top(), nodeBounds.Top()))
		bounds.SetZ(matrix.Max(bounds.Right(), nodeBounds.Right()))
		bounds.SetW(matrix.Max(bounds.Bottom(), nodeBounds.Bottom()))
	}
	return bounds, hasBounds
}

func (g *shaderGraph) zoomValue() matrix.Float {
	if g == nil || g.zoom <= 0 {
		return 1
	}
	return matrix.Clamp(g.zoom, shaderGraphMinZoom, shaderGraphMaxZoom)
}

func (g *shaderGraph) setZoomAroundViewPosition(zoom matrix.Float, anchor matrix.Vec2) {
	if g == nil {
		return
	}
	next := matrix.Clamp(zoom, shaderGraphMinZoom, shaderGraphMaxZoom)
	if matrix.Approx(next, g.zoomValue()) {
		g.zoom = next
		return
	}
	graphAnchor := g.graphPositionFromView(anchor)
	g.zoom = next
	g.pan = anchor.Subtract(graphAnchor.Scale(next))
	g.applyViewOffsets()
}

func (g *shaderGraph) viewPosition(position matrix.Vec2) matrix.Vec2 {
	return position.Scale(g.zoomValue()).Add(g.pan)
}

func (g *shaderGraph) graphPositionFromView(position matrix.Vec2) matrix.Vec2 {
	return position.Subtract(g.pan).Scale(1 / g.zoomValue())
}

func (g *shaderGraph) screenToViewPosition(position matrix.Vec2) matrix.Vec2 {
	if g == nil || g.root == nil {
		return position
	}
	return position.Subtract(g.root.Base().Layout().Offset())
}

func (g *shaderGraph) screenPositionInside(position matrix.Vec2) bool {
	if g == nil || g.root == nil {
		return false
	}
	local := g.screenToViewPosition(position)
	size := g.root.Base().Layout().PixelSize()
	return local.X() >= 0 && local.Y() >= 0 &&
		local.X() <= size.X() && local.Y() <= size.Y()
}
