/******************************************************************************/
/* render_graph_comment.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

const (
	shaderGraphCommentDefaultWidth  = float32(360)
	shaderGraphCommentDefaultHeight = float32(220)
	shaderGraphCommentMinWidth      = matrix.Float(180)
	shaderGraphCommentMinHeight     = matrix.Float(96)
	shaderGraphCommentHeaderHeight  = float32(28)
	shaderGraphCommentGripSize      = float32(18)
	shaderGraphCommentZDepth        = float32(2)
)

var (
	shaderGraphCommentBodyColor   = matrix.NewColor(0.15, 0.16, 0.18, 1)
	shaderGraphCommentHeaderColor = matrix.NewColor(0.18, 0.19, 0.22, 1)
	shaderGraphCommentBorderColor = matrix.NewColor(0.25, 0.27, 0.31, 1)
	shaderGraphCommentGripColor   = matrix.NewColor(0.36, 0.38, 0.43, 1)
)

type shaderGraphComment struct {
	graph          *shaderGraph
	host           *engine.Host
	root           *ui.Panel
	header         *ui.Panel
	labelInput     *ui.Input
	bodyDrag       *ui.Panel
	resizeGrip     *ui.Panel
	selectionFrame *ui.Panel
	id             string
	label          string
	position       matrix.Vec2
	size           matrix.Vec2
	selected       bool
	dragging       bool
	resizing       bool
	dragMouse      matrix.Vec2
	dragOrigin     matrix.Vec2
	sizeOrigin     matrix.Vec2
}

func (c *shaderGraphComment) Initialize(graph *shaderGraph, host *engine.Host, uiMan *ui.Manager, parent *ui.Panel, comment RenderGraphComment) {
	c.graph = graph
	c.host = host
	c.id = comment.ID
	c.label = comment.Label
	if c.label == "" {
		c.label = "Comment"
	}
	c.position = comment.Position
	c.size = shaderGraphCommentSizeOrDefault(comment.Size)

	c.root = uiMan.Add().ToPanel()
	c.root.Init(nil, ui.ElementTypePanel)
	c.root.DontFitContent()
	c.root.SetColor(shaderGraphCommentBodyColor)
	c.root.SetBorderRadius(5, 5, 5, 5)
	c.root.SetBorderSize(1, 1, 1, 1)
	c.root.SetBorderColor(
		shaderGraphCommentBorderColor,
		shaderGraphCommentBorderColor,
		shaderGraphCommentBorderColor,
		shaderGraphCommentBorderColor,
	)
	c.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.root.Base().Layout().SetZ(shaderGraphCommentZDepth)
	c.bindDragEvents(c.root.Base())
	parent.AddChild(c.root.Base())

	c.createHeader(uiMan)
	c.createBodyDragSurface(uiMan)
	c.createResizeGrip(uiMan)
	c.createSelectionFrame(uiMan)
	c.applySize()
	c.applyViewOffset()
}

func (c *shaderGraphComment) Update() {
	if c == nil || (!c.dragging && !c.resizing) || c.host == nil || c.host.Window == nil {
		return
	}
	mouse := &c.host.Window.Mouse
	if !mouse.Held(hid.MouseButtonLeft) && !mouse.Pressed(hid.MouseButtonLeft) {
		c.stopInteraction()
		return
	}
	current := mouse.ScreenPosition()
	delta := current.Subtract(c.dragMouse)
	if c.graph != nil {
		delta = delta.Scale(1 / c.graph.zoomValue())
	}
	if c.resizing {
		c.setSize(c.sizeOrigin.Add(delta))
		return
	}
	if c.dragging {
		c.position = c.dragOrigin.Add(delta)
		c.applyViewOffset()
	}
}

func (c *shaderGraphComment) setSelected(selected bool) {
	if c == nil || c.selected == selected {
		return
	}
	c.selected = selected
	if c.selectionFrame == nil {
		return
	}
	if selected {
		c.selectionFrame.Base().Show()
	} else {
		c.selectionFrame.Base().Hide()
	}
}

func (c *shaderGraphComment) applyViewOffset() {
	if c == nil || c.root == nil {
		return
	}
	position := c.position
	zoom := matrix.Float(1)
	if c.graph != nil {
		zoom = c.graph.zoomValue()
		position = c.graph.viewPosition(position)
	}
	scale := matrix.NewVec3(c.size.X()*zoom, c.size.Y()*zoom, 1)
	if parent := c.root.Base().Entity().Parent; parent != nil {
		parentScale := parent.Transform.WorldScale()
		if !matrix.Approx(parentScale.X(), 0) && !matrix.Approx(parentScale.Y(), 0) {
			scale.SetX(scale.X() / parentScale.X())
			scale.SetY(scale.Y() / parentScale.Y())
		}
	}
	if !c.root.Base().Entity().Transform.Scale().Equals(scale) {
		c.root.Base().Entity().Transform.SetScale(scale)
		c.root.Base().SetDirty(ui.DirtyTypeResize)
	}
	c.root.Base().Layout().SetOffset(position.X(), position.Y())
}

func (c *shaderGraphComment) bounds() matrix.Vec4 {
	if c == nil {
		return matrix.Vec4Zero()
	}
	size := shaderGraphCommentSizeOrDefault(c.size)
	return matrix.NewVec4(
		c.position.X(),
		c.position.Y(),
		c.position.X()+size.X(),
		c.position.Y()+size.Y(),
	)
}

func (c *shaderGraphComment) stopInteraction() {
	if c == nil || (!c.dragging && !c.resizing) {
		return
	}
	wasDragging := c.dragging
	wasResizing := c.resizing
	c.dragging = false
	c.resizing = false
	if c.graph == nil || c.graph.history == nil || c.id == "" {
		return
	}
	if wasDragging && !matrix.Vec2Approx(c.position, c.dragOrigin) {
		c.graph.history.Add(&shaderGraphCommentPositionHistory{
			graph: c.graph,
			id:    c.id,
			from:  c.dragOrigin,
			to:    c.position,
		})
	}
	if wasResizing && !matrix.Vec2Approx(c.size, c.sizeOrigin) {
		c.graph.history.Add(&shaderGraphCommentSizeHistory{
			graph: c.graph,
			id:    c.id,
			from:  c.sizeOrigin,
			to:    c.size,
		})
	}
}

func (c *shaderGraphComment) beginDrag() {
	if c == nil || c.host == nil || c.host.Window == nil {
		return
	}
	if c.graph != nil {
		if c.graph.isPanInputHeld() {
			return
		}
		c.graph.SelectCommentFromInput(c)
	}
	c.dragging = true
	c.resizing = false
	c.dragMouse = c.host.Window.Mouse.ScreenPosition()
	c.dragOrigin = c.position
	c.sizeOrigin = c.size
}

func (c *shaderGraphComment) beginResize() {
	if c == nil || c.host == nil || c.host.Window == nil {
		return
	}
	if c.graph != nil {
		if c.graph.isPanInputHeld() {
			return
		}
		c.graph.SelectCommentFromInput(c)
	}
	c.resizing = true
	c.dragging = false
	c.dragMouse = c.host.Window.Mouse.ScreenPosition()
	c.dragOrigin = c.position
	c.sizeOrigin = c.size
}

func (c *shaderGraphComment) setSize(size matrix.Vec2) {
	if c == nil {
		return
	}
	size = shaderGraphCommentClampSize(size)
	if matrix.Vec2Approx(c.size, size) {
		return
	}
	c.size = size
	c.applySize()
	c.applyViewOffset()
}

func (c *shaderGraphComment) applySize() {
	if c == nil || c.root == nil {
		return
	}
	width := float32(c.size.X())
	height := float32(c.size.Y())
	c.root.Base().Layout().Scale(width, height)
	if c.labelInput != nil {
		c.labelInput.Base().Layout().Scale(max(1, width-16), shaderGraphCommentHeaderHeight)
	}
	if c.header != nil {
		c.header.Base().Layout().Scale(width, shaderGraphCommentHeaderHeight)
	}
	if c.bodyDrag != nil {
		c.bodyDrag.Base().Layout().Scale(width, max(1, height-shaderGraphCommentHeaderHeight))
	}
	if c.resizeGrip != nil {
		c.resizeGrip.Base().Layout().SetOffset(width-shaderGraphCommentGripSize, height-shaderGraphCommentGripSize)
	}
	if c.selectionFrame != nil {
		c.selectionFrame.Base().Layout().Scale(width, height)
	}
}

func (c *shaderGraphComment) bindDragEvents(target *ui.UI) {
	if c == nil || target == nil {
		return
	}
	target.AddEvent(ui.EventTypeDown, c.beginDrag)
	target.AddEvent(ui.EventTypeUp, c.stopInteraction)
	target.AddEvent(ui.EventTypeDragEnd, c.stopInteraction)
}

func (c *shaderGraphComment) createHeader(uiMan *ui.Manager) {
	c.header = uiMan.Add().ToPanel()
	c.header.Init(nil, ui.ElementTypePanel)
	c.header.DontFitContent()
	c.header.SetColor(shaderGraphCommentHeaderColor)
	c.header.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.header.Base().Layout().SetZ(0.1)
	c.header.Base().Layout().Scale(float32(c.size.X()), shaderGraphCommentHeaderHeight)
	c.header.Base().Layout().SetOffset(0, 0)
	c.bindDragEvents(c.header.Base())
	c.root.AddChild(c.header.Base())

	c.labelInput = uiMan.Add().ToInput()
	c.labelInput.Init("Comment")
	c.labelInput.SetTextWithoutEvent(c.label)
	c.labelInput.SetFontSize(12)
	c.labelInput.SetWrap(false)
	c.labelInput.SetFGColor(matrix.NewColor(0.88, 0.90, 0.94, 1))
	c.labelInput.SetBGColor(shaderGraphCommentHeaderColor)
	c.labelInput.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.labelInput.Base().Layout().SetZ(0.2)
	c.labelInput.Base().Layout().SetOffset(6, 0)
	c.labelInput.Base().AddEvent(ui.EventTypeChange, func() {
		c.label = c.labelInput.Text()
	})
	c.header.AddChild(c.labelInput.Base())
}

func (c *shaderGraphComment) createBodyDragSurface(uiMan *ui.Manager) {
	c.bodyDrag = uiMan.Add().ToPanel()
	c.bodyDrag.Init(nil, ui.ElementTypePanel)
	c.bodyDrag.DontFitContent()
	c.bodyDrag.SetColor(matrix.ColorTransparent())
	c.bodyDrag.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.bodyDrag.Base().Layout().SetZ(0.05)
	c.bodyDrag.Base().Layout().SetOffset(0, shaderGraphCommentHeaderHeight)
	c.bindDragEvents(c.bodyDrag.Base())
	c.root.AddChild(c.bodyDrag.Base())
}

func (c *shaderGraphComment) createResizeGrip(uiMan *ui.Manager) {
	c.resizeGrip = uiMan.Add().ToPanel()
	c.resizeGrip.Init(nil, ui.ElementTypePanel)
	c.resizeGrip.DontFitContent()
	c.resizeGrip.SetColor(shaderGraphCommentGripColor)
	c.resizeGrip.SetBorderRadius(3, 3, 3, 3)
	c.resizeGrip.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.resizeGrip.Base().Layout().SetZ(0.4)
	c.resizeGrip.Base().Layout().Scale(shaderGraphCommentGripSize, shaderGraphCommentGripSize)
	c.resizeGrip.Base().AddEvent(ui.EventTypeDown, c.beginResize)
	c.resizeGrip.Base().AddEvent(ui.EventTypeUp, c.stopInteraction)
	c.resizeGrip.Base().AddEvent(ui.EventTypeDragEnd, c.stopInteraction)
	c.root.AddChild(c.resizeGrip.Base())
}

func (c *shaderGraphComment) createSelectionFrame(uiMan *ui.Manager) {
	c.selectionFrame = uiMan.Add().ToPanel()
	c.selectionFrame.Init(nil, ui.ElementTypePanel)
	c.selectionFrame.AllowClickThrough()
	c.selectionFrame.DontFitContent()
	c.selectionFrame.SetColor(matrix.ColorTransparent())
	c.selectionFrame.SetBorderRadius(5, 5, 5, 5)
	c.selectionFrame.SetBorderSize(2, 2, 2, 2)
	c.selectionFrame.SetBorderColor(
		shaderGraphNodeSelectColor,
		shaderGraphNodeSelectColor,
		shaderGraphNodeSelectColor,
		shaderGraphNodeSelectColor,
	)
	c.selectionFrame.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.selectionFrame.Base().Layout().SetZ(0.8)
	c.selectionFrame.Base().Layout().SetOffset(0, 0)
	c.selectionFrame.Base().Hide()
	c.root.AddChild(c.selectionFrame.Base())
}

func shaderGraphCommentSizeOrDefault(size matrix.Vec2) matrix.Vec2 {
	if size.X() <= 0 {
		size.SetX(matrix.Float(shaderGraphCommentDefaultWidth))
	}
	if size.Y() <= 0 {
		size.SetY(matrix.Float(shaderGraphCommentDefaultHeight))
	}
	return shaderGraphCommentClampSize(size)
}

func shaderGraphCommentClampSize(size matrix.Vec2) matrix.Vec2 {
	if size.X() < shaderGraphCommentMinWidth {
		size.SetX(shaderGraphCommentMinWidth)
	}
	if size.Y() < shaderGraphCommentMinHeight {
		size.SetY(shaderGraphCommentMinHeight)
	}
	return size
}
