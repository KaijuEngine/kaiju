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
	renderGraphCommentDefaultWidth  = matrix.Float(360)
	renderGraphCommentDefaultHeight = matrix.Float(220)
	renderGraphCommentMinWidth      = matrix.Float(180)
	renderGraphCommentMinHeight     = matrix.Float(96)
	renderGraphCommentHeaderHeight  = matrix.Float(28)
	renderGraphCommentGripSize      = matrix.Float(18)
	renderGraphCommentZDepth        = matrix.Float(2)
)

var (
	renderGraphCommentBodyColor   = matrix.NewColor(0.15, 0.16, 0.18, 1)
	renderGraphCommentHeaderColor = matrix.NewColor(0.18, 0.19, 0.22, 1)
	renderGraphCommentBorderColor = matrix.NewColor(0.25, 0.27, 0.31, 1)
	renderGraphCommentGripColor   = matrix.NewColor(0.36, 0.38, 0.43, 1)
)

type renderGraphComment struct {
	graph          *renderGraph
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

func (c *renderGraphComment) Initialize(graph *renderGraph, host *engine.Host, uiMan *ui.Manager, parent *ui.Panel, comment RenderGraphComment) {
	c.graph = graph
	c.host = host
	c.id = comment.ID
	c.label = comment.Label
	if c.label == "" {
		c.label = "Comment"
	}
	c.position = comment.Position
	c.size = renderGraphCommentSizeOrDefault(comment.Size)

	c.root = uiMan.Add().ToPanel()
	c.root.Init(nil, ui.ElementTypePanel)
	c.root.DontFitContent()
	c.root.SetColor(renderGraphCommentBodyColor)
	c.root.SetBorderRadius(5, 5, 5, 5)
	c.root.SetBorderSize(1, 1, 1, 1)
	c.root.SetBorderColor(
		renderGraphCommentBorderColor,
		renderGraphCommentBorderColor,
		renderGraphCommentBorderColor,
		renderGraphCommentBorderColor,
	)
	c.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.root.Base().Layout().SetZ(renderGraphCommentZDepth)
	c.bindDragEvents(c.root.Base())
	parent.AddChild(c.root.Base())

	c.createHeader(uiMan)
	c.createBodyDragSurface(uiMan)
	c.createResizeGrip(uiMan)
	c.createSelectionFrame(uiMan)
	c.applySize()
	c.applyViewOffset()
}

func (c *renderGraphComment) Update() {
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

func (c *renderGraphComment) setSelected(selected bool) {
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

func (c *renderGraphComment) applyViewOffset() {
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

func (c *renderGraphComment) bounds() matrix.Vec4 {
	if c == nil {
		return matrix.Vec4Zero()
	}
	size := renderGraphCommentSizeOrDefault(c.size)
	return matrix.NewVec4(
		c.position.X(),
		c.position.Y(),
		c.position.X()+size.X(),
		c.position.Y()+size.Y(),
	)
}

func (c *renderGraphComment) stopInteraction() {
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
		c.graph.history.Add(&renderGraphCommentPositionHistory{
			graph: c.graph,
			id:    c.id,
			from:  c.dragOrigin,
			to:    c.position,
		})
	}
	if wasResizing && !matrix.Vec2Approx(c.size, c.sizeOrigin) {
		c.graph.history.Add(&renderGraphCommentSizeHistory{
			graph: c.graph,
			id:    c.id,
			from:  c.sizeOrigin,
			to:    c.size,
		})
	}
}

func (c *renderGraphComment) beginDrag() {
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

func (c *renderGraphComment) beginResize() {
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

func (c *renderGraphComment) setSize(size matrix.Vec2) {
	if c == nil {
		return
	}
	size = renderGraphCommentClampSize(size)
	if matrix.Vec2Approx(c.size, size) {
		return
	}
	c.size = size
	c.applySize()
	c.applyViewOffset()
}

func (c *renderGraphComment) applySize() {
	if c == nil || c.root == nil {
		return
	}
	width := matrix.Float(c.size.X())
	height := matrix.Float(c.size.Y())
	c.root.Base().Layout().Scale(width, height)
	if c.labelInput != nil {
		c.labelInput.Base().Layout().Scale(max(1, width-16), renderGraphCommentHeaderHeight)
	}
	if c.header != nil {
		c.header.Base().Layout().Scale(width, renderGraphCommentHeaderHeight)
	}
	if c.bodyDrag != nil {
		c.bodyDrag.Base().Layout().Scale(width, max(1, height-renderGraphCommentHeaderHeight))
	}
	if c.resizeGrip != nil {
		c.resizeGrip.Base().Layout().SetOffset(width-renderGraphCommentGripSize, height-renderGraphCommentGripSize)
	}
	if c.selectionFrame != nil {
		c.selectionFrame.Base().Layout().Scale(width, height)
	}
}

func (c *renderGraphComment) bindDragEvents(target *ui.UI) {
	if c == nil || target == nil {
		return
	}
	target.AddEvent(ui.EventTypeDown, c.beginDrag)
	target.AddEvent(ui.EventTypeUp, c.stopInteraction)
	target.AddEvent(ui.EventTypeDragEnd, c.stopInteraction)
}

func (c *renderGraphComment) createHeader(uiMan *ui.Manager) {
	c.header = uiMan.Add().ToPanel()
	c.header.Init(nil, ui.ElementTypePanel)
	c.header.DontFitContent()
	c.header.SetColor(renderGraphCommentHeaderColor)
	c.header.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.header.Base().Layout().SetZ(0.1)
	c.header.Base().Layout().Scale(matrix.Float(c.size.X()), renderGraphCommentHeaderHeight)
	c.header.Base().Layout().SetOffset(0, 0)
	c.bindDragEvents(c.header.Base())
	c.root.AddChild(c.header.Base())

	c.labelInput = uiMan.Add().ToInput()
	c.labelInput.Init("Comment")
	c.labelInput.SetTextWithoutEvent(c.label)
	c.labelInput.SetFontSize(12)
	c.labelInput.SetWrap(false)
	c.labelInput.SetFGColor(matrix.NewColor(0.88, 0.90, 0.94, 1))
	c.labelInput.SetBGColor(renderGraphCommentHeaderColor)
	c.labelInput.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.labelInput.Base().Layout().SetZ(0.2)
	c.labelInput.Base().Layout().SetOffset(6, 0)
	c.labelInput.Base().AddEvent(ui.EventTypeChange, func() {
		c.label = c.labelInput.Text()
	})
	c.header.AddChild(c.labelInput.Base())
}

func (c *renderGraphComment) createBodyDragSurface(uiMan *ui.Manager) {
	c.bodyDrag = uiMan.Add().ToPanel()
	c.bodyDrag.Init(nil, ui.ElementTypePanel)
	c.bodyDrag.DontFitContent()
	c.bodyDrag.SetColor(matrix.ColorTransparent())
	c.bodyDrag.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.bodyDrag.Base().Layout().SetZ(0.05)
	c.bodyDrag.Base().Layout().SetOffset(0, renderGraphCommentHeaderHeight)
	c.bindDragEvents(c.bodyDrag.Base())
	c.root.AddChild(c.bodyDrag.Base())
}

func (c *renderGraphComment) createResizeGrip(uiMan *ui.Manager) {
	c.resizeGrip = uiMan.Add().ToPanel()
	c.resizeGrip.Init(nil, ui.ElementTypePanel)
	c.resizeGrip.DontFitContent()
	c.resizeGrip.SetColor(renderGraphCommentGripColor)
	c.resizeGrip.SetBorderRadius(3, 3, 3, 3)
	c.resizeGrip.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.resizeGrip.Base().Layout().SetZ(0.4)
	c.resizeGrip.Base().Layout().Scale(renderGraphCommentGripSize, renderGraphCommentGripSize)
	c.resizeGrip.Base().AddEvent(ui.EventTypeDown, c.beginResize)
	c.resizeGrip.Base().AddEvent(ui.EventTypeUp, c.stopInteraction)
	c.resizeGrip.Base().AddEvent(ui.EventTypeDragEnd, c.stopInteraction)
	c.root.AddChild(c.resizeGrip.Base())
}

func (c *renderGraphComment) createSelectionFrame(uiMan *ui.Manager) {
	c.selectionFrame = uiMan.Add().ToPanel()
	c.selectionFrame.Init(nil, ui.ElementTypePanel)
	c.selectionFrame.AllowClickThrough()
	c.selectionFrame.DontFitContent()
	c.selectionFrame.SetColor(matrix.ColorTransparent())
	c.selectionFrame.SetBorderRadius(5, 5, 5, 5)
	c.selectionFrame.SetBorderSize(2, 2, 2, 2)
	c.selectionFrame.SetBorderColor(
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
	)
	c.selectionFrame.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	c.selectionFrame.Base().Layout().SetZ(0.8)
	c.selectionFrame.Base().Layout().SetOffset(0, 0)
	c.selectionFrame.Base().Hide()
	c.root.AddChild(c.selectionFrame.Base())
}

func renderGraphCommentSizeOrDefault(size matrix.Vec2) matrix.Vec2 {
	if size.X() <= 0 {
		size.SetX(matrix.Float(renderGraphCommentDefaultWidth))
	}
	if size.Y() <= 0 {
		size.SetY(matrix.Float(renderGraphCommentDefaultHeight))
	}
	return renderGraphCommentClampSize(size)
}

func renderGraphCommentClampSize(size matrix.Vec2) matrix.Vec2 {
	if size.X() < renderGraphCommentMinWidth {
		size.SetX(renderGraphCommentMinWidth)
	}
	if size.Y() < renderGraphCommentMinHeight {
		size.SetY(renderGraphCommentMinHeight)
	}
	return size
}
