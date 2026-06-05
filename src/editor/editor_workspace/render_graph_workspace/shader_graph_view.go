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

func (g *shaderGraph) stopNodeDrags() {
	for i := range g.nodes {
		g.nodes[i].stopDrag()
	}
}

func (g *shaderGraph) applyViewOffsets() {
	for i := range g.nodes {
		g.nodes[i].applyViewOffset()
	}
}

func (g *shaderGraph) CenterView() {
	if g == nil {
		return
	}
	g.pan = matrix.Vec2Zero()
	g.applyViewOffsets()
}

func (g *shaderGraph) viewPosition(position matrix.Vec2) matrix.Vec2 {
	return position.Add(g.pan)
}

func (g *shaderGraph) graphPositionFromView(position matrix.Vec2) matrix.Vec2 {
	return position.Subtract(g.pan)
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
