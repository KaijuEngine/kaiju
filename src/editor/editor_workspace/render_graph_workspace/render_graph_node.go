/******************************************************************************/
/* render_graph_node.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package render_graph_workspace

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/rendering"
)

const (
	renderGraphNodeWidth        = float32(210)
	renderGraphNodeBaseHeight   = float32(82)
	renderGraphNodePortHeight   = float32(20)
	renderGraphNodeHeaderH      = float32(24)
	renderGraphNodePadding      = float32(12)
	renderGraphNodePortLabelGap = float32(8)
	renderGraphNodeFieldStartY  = renderGraphNodeHeaderH + 38
	renderGraphNodePortStartY   = renderGraphNodeHeaderH + 42
	renderGraphNodePortDotSize  = float32(10)
)

var (
	renderGraphNodeAccentColor = matrix.NewColor(0.4078, 0.1647, 0.1765, 1) // #682A2D from default.css
	renderGraphNodeBodyColor   = matrix.NewColor(0.12, 0.13, 0.15, 1)
	renderGraphNodeSelectColor = matrix.NewColor(0.95, 0.72, 0.28, 1)
	renderGraphSurfaceColor    = matrix.NewColor(0.39, 0.82, 0.43, 1.0)
)

type renderGraphNode struct {
	graph          *renderGraph
	host           *engine.Host
	root           *ui.Panel
	bodyDrag       *ui.Panel
	selectionFrame *ui.Panel
	id             string
	typeID         string
	title          *ui.Label
	description    *ui.Label
	fields         []*renderGraphNodeField
	inputs         []*renderGraphPort
	outputs        []*renderGraphPort
	values         map[string]renderGraphNodeFieldValue
	position       matrix.Vec2
	height         float32
	zDepth         float32
	zSlot          int
	zSlotAssigned  bool
	selected       bool
	dragging       bool
	dragMouse      matrix.Vec2
	dragOrigin     matrix.Vec2
	dragNodes      []*renderGraphNode
	dragOrigins    []matrix.Vec2
}

func (n *renderGraphNode) Initialize(graph *renderGraph, host *engine.Host, uiMan *ui.Manager, parent *ui.Panel, spec renderGraphNodeSpec, position matrix.Vec2) {
	n.graph = graph
	n.host = host
	n.position = position
	n.values = make(map[string]renderGraphNodeFieldValue, len(spec.Fields))
	n.root = uiMan.Add().ToPanel()
	n.root.Init(nil, ui.ElementTypePanel)
	n.root.DontFitContent()
	n.root.SetColor(renderGraphNodeBodyColor)
	n.root.SetBorderRadius(6, 6, 6, 6)
	n.root.SetBorderSize(1, 1, 1, 1)
	n.root.SetBorderColor(
		renderGraphNodeAccentColor,
		renderGraphNodeAccentColor,
		renderGraphNodeAccentColor,
		renderGraphNodeAccentColor,
	)
	n.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.root.Base().Layout().SetZ(n.zDepth)
	n.bindDragEvents(n.root.Base())
	parent.AddChild(n.root.Base())

	height := renderGraphNodeHeight(spec)
	n.height = height
	n.root.Base().Layout().Scale(renderGraphNodeWidth, height)
	n.applyViewOffset()

	n.createBodyDragSurface(uiMan, height)
	n.createHeader(uiMan, spec.Name)
	n.createDescription(uiMan, spec.Description)
	n.createFields(uiMan, spec.Fields)
	n.createPorts(uiMan, spec)
	n.createSelectionFrame(uiMan, height)
}

func (n *renderGraphNode) Input(index int) *renderGraphPort {
	if index < 0 || index >= len(n.inputs) {
		return nil
	}
	return n.inputs[index]
}

func (n *renderGraphNode) Output(index int) *renderGraphPort {
	if index < 0 || index >= len(n.outputs) {
		return nil
	}
	return n.outputs[index]
}

func (n *renderGraphNode) FieldValue(id string) renderGraphNodeFieldValue {
	if n == nil || n.values == nil {
		return renderGraphNodeFieldValue{}
	}
	return n.values[id].Clone()
}

func (n *renderGraphNode) setFieldValue(id string, value renderGraphNodeFieldValue) {
	if n == nil || n.values == nil || id == "" {
		return
	}
	n.values[id] = value.Clone()
}

func (n *renderGraphNode) Update() {
	if !n.dragging || n.host == nil || n.host.Window == nil {
		return
	}
	mouse := &n.host.Window.Mouse
	if !mouse.Held(hid.MouseButtonLeft) && !mouse.Pressed(hid.MouseButtonLeft) {
		n.stopDrag()
		return
	}
	current := mouse.ScreenPosition()
	delta := current.Subtract(n.dragMouse)
	if n.graph != nil {
		delta = delta.Scale(1 / n.graph.zoomValue())
	}
	n.applyDragDelta(delta)
}

func (n *renderGraphNode) beginDrag() {
	if n.host == nil || n.host.Window == nil {
		return
	}
	if n.graph != nil && n.graph.isPanInputHeld() {
		return
	}
	n.dragging = true
	n.dragMouse = n.host.Window.Mouse.ScreenPosition()
	n.dragOrigin = n.position
	n.captureDragNodes()
}

func (n *renderGraphNode) setSelected(selected bool) {
	if n == nil || n.selected == selected {
		return
	}
	n.selected = selected
	if n.root == nil {
		return
	}
	if selected {
		if n.selectionFrame != nil {
			n.selectionFrame.Base().Show()
		}
		return
	}
	if n.selectionFrame != nil {
		n.selectionFrame.Base().Hide()
	}
}

func (n *renderGraphNode) stopDrag() {
	if !n.dragging {
		return
	}
	n.addDragHistory()
	n.dragging = false
	n.dragNodes = nil
	n.dragOrigins = nil
}

func (n *renderGraphNode) applyViewOffset() {
	if n.root == nil {
		return
	}
	position := n.position
	zoom := matrix.Float(1)
	if n.graph != nil {
		zoom = n.graph.zoomValue()
		position = n.graph.viewPosition(position)
	}
	scale := matrix.NewVec3(
		matrix.Float(renderGraphNodeWidth)*zoom,
		matrix.Float(n.height)*zoom,
		1,
	)
	if parent := n.root.Base().Entity().Parent; parent != nil {
		parentScale := parent.Transform.WorldScale()
		if !matrix.Approx(parentScale.X(), 0) && !matrix.Approx(parentScale.Y(), 0) {
			scale.SetX(scale.X() / parentScale.X())
			scale.SetY(scale.Y() / parentScale.Y())
		}
	}
	if !n.root.Base().Entity().Transform.Scale().Equals(scale) {
		n.root.Base().Entity().Transform.SetScale(scale)
		n.root.Base().SetDirty(ui.DirtyTypeResize)
	}
	n.root.Base().Layout().SetOffset(position.X(), position.Y())
}

func (n *renderGraphNode) bindDragEvents(target *ui.UI) {
	n.bindSelectionEvent(target)
	target.AddEvent(ui.EventTypeDown, n.beginDrag)
	target.AddEvent(ui.EventTypeUp, n.stopDrag)
	target.AddEvent(ui.EventTypeDragEnd, n.stopDrag)
}

func (n *renderGraphNode) captureDragNodes() {
	nodes := []*renderGraphNode{n}
	if n.graph != nil && n.graph.IsSelected(n) {
		selection := n.graph.Selection()
		if len(selection) > 0 {
			nodes = selection
		}
	}
	n.dragNodes = nodes
	n.dragOrigins = make([]matrix.Vec2, len(nodes))
	for i := range nodes {
		if nodes[i] != nil {
			n.dragOrigins[i] = nodes[i].position
		}
	}
}

func (n *renderGraphNode) applyDragDelta(delta matrix.Vec2) {
	if len(n.dragNodes) == 0 {
		n.position = n.dragOrigin.Add(delta)
		n.applyViewOffset()
		return
	}
	for i := range n.dragNodes {
		node := n.dragNodes[i]
		if node == nil || i >= len(n.dragOrigins) {
			continue
		}
		node.position = n.dragOrigins[i].Add(delta)
		node.applyViewOffset()
	}
}

func (n *renderGraphNode) addDragHistory() {
	if n.graph == nil || n.graph.history == nil || len(n.dragNodes) == 0 {
		return
	}
	history := &renderGraphNodePositionHistory{
		graph: n.graph,
		ids:   make([]string, 0, len(n.dragNodes)),
		from:  make([]matrix.Vec2, 0, len(n.dragNodes)),
		to:    make([]matrix.Vec2, 0, len(n.dragNodes)),
	}
	for i := range n.dragNodes {
		node := n.dragNodes[i]
		if node == nil || node.id == "" || i >= len(n.dragOrigins) {
			continue
		}
		if matrix.Vec2Approx(node.position, n.dragOrigins[i]) {
			continue
		}
		history.ids = append(history.ids, node.id)
		history.from = append(history.from, n.dragOrigins[i])
		history.to = append(history.to, node.position)
	}
	if len(history.ids) > 0 {
		n.graph.history.Add(history)
	}
}

func (n *renderGraphNode) createBodyDragSurface(uiMan *ui.Manager, height float32) {
	n.bodyDrag = uiMan.Add().ToPanel()
	n.bodyDrag.Init(nil, ui.ElementTypePanel)
	n.bodyDrag.DontFitContent()
	n.bodyDrag.SetColor(matrix.ColorTransparent())
	n.bodyDrag.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.bodyDrag.Base().Layout().SetZ(0.05)
	n.bodyDrag.Base().Layout().Scale(renderGraphNodeWidth, max(1, height-renderGraphNodeHeaderH))
	n.bodyDrag.Base().Layout().SetOffset(0, renderGraphNodeHeaderH)
	n.bindDragEvents(n.bodyDrag.Base())
	n.root.AddChild(n.bodyDrag.Base())
}

func (n *renderGraphNode) createHeader(uiMan *ui.Manager, name string) {
	header := uiMan.Add().ToPanel()
	header.Init(nil, ui.ElementTypePanel)
	header.DontFitContent()
	header.SetColor(renderGraphNodeAccentColor)
	header.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	header.Base().Layout().SetZ(0.1)
	header.Base().Layout().Scale(renderGraphNodeWidth, renderGraphNodeHeaderH)
	header.Base().Layout().SetOffset(0, 0)
	n.bindDragEvents(header.Base())
	n.root.AddChild(header.Base())

	n.title = uiMan.Add().ToLabel()
	n.title.Init(name)
	n.title.SetFontSize(12)
	n.title.SetWrap(false)
	n.title.SetColor(matrix.ColorWhite())
	n.title.SetBaseline(rendering.FontBaselineCenter)
	n.title.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.title.Base().Layout().SetZ(0.2)
	n.title.Base().Layout().Scale(renderGraphNodeWidth-renderGraphNodePadding*2, renderGraphNodeHeaderH)
	n.title.Base().Layout().SetOffset(renderGraphNodePadding, 0)
	header.AddChild(n.title.Base())
}

func (n *renderGraphNode) createSelectionFrame(uiMan *ui.Manager, height float32) {
	n.selectionFrame = uiMan.Add().ToPanel()
	n.selectionFrame.Init(nil, ui.ElementTypePanel)
	n.selectionFrame.AllowClickThrough()
	n.selectionFrame.DontFitContent()
	n.selectionFrame.SetColor(matrix.ColorTransparent())
	n.selectionFrame.SetBorderRadius(6, 6, 6, 6)
	n.selectionFrame.SetBorderSize(2, 2, 2, 2)
	n.selectionFrame.SetBorderColor(
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
		renderGraphNodeSelectColor,
	)
	n.selectionFrame.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.selectionFrame.Base().Layout().SetZ(1)
	n.selectionFrame.Base().Layout().Scale(renderGraphNodeWidth, height)
	n.selectionFrame.Base().Layout().SetOffset(0, 0)
	n.selectionFrame.Base().Hide()
	n.root.AddChild(n.selectionFrame.Base())
}

func (n *renderGraphNode) createDescription(uiMan *ui.Manager, description string) {
	n.description = uiMan.Add().ToLabel()
	n.description.Init(description)
	n.description.SetFontSize(9)
	n.description.SetColor(matrix.NewColor(0.70, 0.74, 0.80, 1))
	n.description.SetWidthAutoHeight(renderGraphNodeWidth - renderGraphNodePadding*2)
	n.description.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.description.Base().Layout().SetZ(0.1)
	n.description.Base().Layout().SetOffset(renderGraphNodePadding, renderGraphNodeHeaderH+6)
	n.root.AddChild(n.description.Base())
}

func (n *renderGraphNode) createPorts(uiMan *ui.Manager, spec renderGraphNodeSpec) {
	inputs := spec.Inputs
	outputs := spec.Outputs
	rowCount := max(len(inputs), len(outputs))
	startY := renderGraphNodePortStartY + renderGraphNodeFieldsHeight(spec.Fields)
	for i := range rowCount {
		y := startY + float32(i)*renderGraphNodePortHeight
		if i < len(inputs) {
			n.inputs = append(n.inputs, n.createPort(uiMan, inputs[i], false, i, y, i < len(outputs)))
		}
		if i < len(outputs) {
			n.outputs = append(n.outputs, n.createPort(uiMan, outputs[i], true, i, y, i < len(inputs)))
		}
	}
}

func (n *renderGraphNode) createPort(uiMan *ui.Manager, port renderGraphPortSpec, output bool, index int, y float32, paired bool) *renderGraphPort {
	const dotSize = renderGraphNodePortDotSize
	dotX := renderGraphNodePadding
	dotY := y + (renderGraphNodePortHeight-dotSize)*0.5
	labelX := renderGraphNodePadding + dotSize + renderGraphNodePortLabelGap
	labelWidth := renderGraphNodeWidth*0.5 - renderGraphNodePadding*2 - dotSize - renderGraphNodePortLabelGap
	justify := rendering.FontJustifyLeft
	if output {
		dotX = renderGraphNodeWidth - renderGraphNodePadding - dotSize
		if paired {
			labelX = renderGraphNodeWidth*0.5 - renderGraphNodePadding
			labelWidth = dotX - labelX - renderGraphNodePortLabelGap
		} else {
			labelWidth = renderGraphNodeWidth * 0.45
			labelX = dotX - labelWidth - renderGraphNodePortLabelGap
		}
	}
	if !paired {
		if !output {
			labelWidth = renderGraphNodeWidth - renderGraphNodePadding*2 - dotSize - renderGraphNodePortLabelGap
		}
	}

	hitX := min(dotX, labelX)
	hitRight := max(dotX+dotSize, labelX+labelWidth)
	hit := uiMan.Add().ToPanel()
	hit.Init(nil, ui.ElementTypePanel)
	hit.DontFitContent()
	hit.SetColor(matrix.ColorTransparent())
	hit.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	hit.Base().Layout().SetZ(0.35)
	hit.Base().Layout().Scale(hitRight-hitX, renderGraphNodePortHeight)
	hit.Base().Layout().SetOffset(hitX, y)
	n.bindSelectionEvent(hit.Base())
	n.root.AddChild(hit.Base())

	dot := uiMan.Add().ToPanel()
	dot.Init(nil, ui.ElementTypePanel)
	dot.DontFitContent()
	dot.SetColor(renderGraphPortColor(port.Type, output))
	dot.SetBorderRadius(dotSize*0.5, dotSize*0.5, dotSize*0.5, dotSize*0.5)
	dot.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	dot.Base().Layout().SetZ(0.2)
	dot.Base().Layout().Scale(dotSize, dotSize)
	dot.Base().Layout().SetOffset(dotX, dotY)
	n.bindSelectionEvent(dot.Base())
	n.root.AddChild(dot.Base())

	label := uiMan.Add().ToLabel()
	label.Init(renderGraphPortLabel(port, output))
	label.SetFontSize(10)
	label.SetWrap(false)
	label.SetColor(matrix.NewColor(0.86, 0.88, 0.91, 1))
	label.SetJustify(justify)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(0.2)
	label.Base().Layout().Scale(labelWidth, renderGraphNodePortHeight)
	label.Base().Layout().SetOffset(labelX, y)
	n.bindSelectionEvent(label.Base())
	n.root.AddChild(label.Base())

	graphPort := &renderGraphPort{
		graph:       n.graph,
		node:        n,
		spec:        port,
		output:      output,
		index:       index,
		hit:         hit,
		dot:         dot,
		label:       label,
		localAnchor: matrix.NewVec2(matrix.Float(dotX+dotSize*0.5), matrix.Float(dotY+dotSize*0.5)),
	}
	graphPort.bindEvents()
	return graphPort
}

func renderGraphNodeHeight(spec renderGraphNodeSpec) float32 {
	ports := max(len(spec.Inputs), len(spec.Outputs))
	return renderGraphNodeBaseHeight + renderGraphNodeFieldsHeight(spec.Fields) +
		float32(ports)*renderGraphNodePortHeight
}

func renderGraphNodeFieldsHeight(fields []renderGraphNodeFieldSpec) float32 {
	if len(fields) == 0 {
		return 0
	}
	height := float32(4)
	for i := range fields {
		height += renderGraphNodeFieldHeight(fields[i]) + renderGraphFieldGap
	}
	return height
}

func renderGraphPortLabel(port renderGraphPortSpec, output bool) string {
	if output {
		return port.Name
	}
	if port.Type == "" {
		return port.Name
	}
	return fmt.Sprintf("%s  %s", port.Name, port.Type)
}

func renderGraphPortColor(portType string, output bool) matrix.Color {
	switch portType {
	case "float":
		return matrix.NewColor(0.35, 0.62, 0.92, 1)
	case "vec2":
		return matrix.NewColor(0.46, 0.73, 0.86, 1)
	case "vec3":
		return matrix.NewColor(0.42, 0.76, 0.47, 1)
	case "vec4":
		return matrix.NewColor(0.58, 0.55, 0.90, 1)
	case "color":
		return matrix.NewColor(0.91, 0.58, 0.30, 1)
	case "texture2D", "texture2d":
		return matrix.NewColor(0.78, 0.61, 0.35, 1)
	case "surface", "volume":
		return renderGraphSurfaceColor
	default:
		if output {
			return matrix.NewColor(0.76, 0.58, 0.92, 1)
		}
		return matrix.NewColor(0.58, 0.64, 0.72, 1)
	}
}
