/******************************************************************************/
/* shader_graph_node.go                                                       */
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
	shaderGraphNodeWidth       = float32(210)
	shaderGraphNodeBaseHeight  = float32(74)
	shaderGraphNodePortHeight  = float32(20)
	shaderGraphNodeHeaderH     = float32(24)
	shaderGraphNodePadding     = float32(8)
	shaderGraphNodeFieldStartY = shaderGraphNodeHeaderH + 38
	shaderGraphNodePortStartY  = shaderGraphNodeHeaderH + 42
	shaderGraphNodePortDotSize = float32(7)
)

var (
	shaderGraphNodeAccentColor = matrix.NewColor(0.4078, 0.1647, 0.1765, 1) // #682A2D from default.css
	shaderGraphNodeBodyColor   = matrix.NewColor(0.12, 0.13, 0.15, 1)
	shaderGraphNodeSelectColor = matrix.NewColor(0.95, 0.72, 0.28, 1)
	shaderGraphSurfaceColor    = matrix.NewColor(0.39, 0.82, 0.43, 1.0)
)

type shaderGraphNode struct {
	graph       *shaderGraph
	host        *engine.Host
	root        *ui.Panel
	bodyDrag    *ui.Panel
	id          string
	typeID      string
	title       *ui.Label
	description *ui.Label
	fields      []*shaderGraphNodeField
	inputs      []*shaderGraphPort
	outputs     []*shaderGraphPort
	values      map[string]shaderGraphNodeFieldValue
	position    matrix.Vec2
	height      float32
	selected    bool
	dragging    bool
	dragMouse   matrix.Vec2
	dragOrigin  matrix.Vec2
	dragNodes   []*shaderGraphNode
	dragOrigins []matrix.Vec2
}

func (n *shaderGraphNode) Initialize(graph *shaderGraph, host *engine.Host, uiMan *ui.Manager, parent *ui.Panel, spec shaderGraphNodeSpec, position matrix.Vec2) {
	n.graph = graph
	n.host = host
	n.position = position
	n.values = make(map[string]shaderGraphNodeFieldValue, len(spec.Fields))
	n.root = uiMan.Add().ToPanel()
	n.root.Init(nil, ui.ElementTypePanel)
	n.root.DontFitContent()
	n.root.SetColor(shaderGraphNodeBodyColor)
	n.root.SetBorderRadius(6, 6, 6, 6)
	n.root.SetBorderSize(1, 1, 1, 1)
	n.root.SetBorderColor(
		shaderGraphNodeAccentColor,
		shaderGraphNodeAccentColor,
		shaderGraphNodeAccentColor,
		shaderGraphNodeAccentColor,
	)
	n.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.root.Base().Layout().SetZ(5)
	n.bindDragEvents(n.root.Base())
	parent.AddChild(n.root.Base())

	height := shaderGraphNodeHeight(spec)
	n.height = height
	n.root.Base().Layout().Scale(shaderGraphNodeWidth, height)
	n.applyViewOffset()

	n.createBodyDragSurface(uiMan, height)
	n.createHeader(uiMan, spec.Name)
	n.createDescription(uiMan, spec.Description)
	n.createFields(uiMan, spec.Fields)
	n.createPorts(uiMan, spec)
}

func (n *shaderGraphNode) Input(index int) *shaderGraphPort {
	if index < 0 || index >= len(n.inputs) {
		return nil
	}
	return n.inputs[index]
}

func (n *shaderGraphNode) Output(index int) *shaderGraphPort {
	if index < 0 || index >= len(n.outputs) {
		return nil
	}
	return n.outputs[index]
}

func (n *shaderGraphNode) FieldValue(id string) shaderGraphNodeFieldValue {
	if n == nil || n.values == nil {
		return shaderGraphNodeFieldValue{}
	}
	return n.values[id]
}

func (n *shaderGraphNode) setFieldValue(id string, value shaderGraphNodeFieldValue) {
	if n == nil || n.values == nil || id == "" {
		return
	}
	n.values[id] = value
}

func (n *shaderGraphNode) Update() {
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

func (n *shaderGraphNode) beginDrag() {
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

func (n *shaderGraphNode) setSelected(selected bool) {
	if n == nil || n.selected == selected {
		return
	}
	n.selected = selected
	if n.root == nil {
		return
	}
	if selected {
		n.root.SetBorderSize(2, 2, 2, 2)
		n.root.SetBorderColor(
			shaderGraphNodeSelectColor,
			shaderGraphNodeSelectColor,
			shaderGraphNodeSelectColor,
			shaderGraphNodeSelectColor,
		)
		n.root.Base().SetDirty(ui.DirtyTypeColorChange)
		return
	}
	n.root.SetBorderSize(1, 1, 1, 1)
	n.root.SetBorderColor(
		shaderGraphNodeAccentColor,
		shaderGraphNodeAccentColor,
		shaderGraphNodeAccentColor,
		shaderGraphNodeAccentColor,
	)
	n.root.Base().SetDirty(ui.DirtyTypeColorChange)
}

func (n *shaderGraphNode) stopDrag() {
	if !n.dragging {
		return
	}
	n.addDragHistory()
	n.dragging = false
	n.dragNodes = nil
	n.dragOrigins = nil
}

func (n *shaderGraphNode) applyViewOffset() {
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
		matrix.Float(shaderGraphNodeWidth)*zoom,
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

func (n *shaderGraphNode) bindDragEvents(target *ui.UI) {
	n.bindSelectionEvent(target)
	target.AddEvent(ui.EventTypeDown, n.beginDrag)
	target.AddEvent(ui.EventTypeUp, n.stopDrag)
	target.AddEvent(ui.EventTypeDragEnd, n.stopDrag)
}

func (n *shaderGraphNode) captureDragNodes() {
	nodes := []*shaderGraphNode{n}
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

func (n *shaderGraphNode) applyDragDelta(delta matrix.Vec2) {
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

func (n *shaderGraphNode) addDragHistory() {
	if n.graph == nil || n.graph.history == nil || len(n.dragNodes) == 0 {
		return
	}
	history := &shaderGraphNodePositionHistory{
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

func (n *shaderGraphNode) createBodyDragSurface(uiMan *ui.Manager, height float32) {
	n.bodyDrag = uiMan.Add().ToPanel()
	n.bodyDrag.Init(nil, ui.ElementTypePanel)
	n.bodyDrag.DontFitContent()
	n.bodyDrag.SetColor(matrix.ColorTransparent())
	n.bodyDrag.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.bodyDrag.Base().Layout().SetZ(5.05)
	n.bodyDrag.Base().Layout().Scale(shaderGraphNodeWidth, max(1, height-shaderGraphNodeHeaderH))
	n.bodyDrag.Base().Layout().SetOffset(0, shaderGraphNodeHeaderH)
	n.bindDragEvents(n.bodyDrag.Base())
	n.root.AddChild(n.bodyDrag.Base())
}

func (n *shaderGraphNode) createHeader(uiMan *ui.Manager, name string) {
	header := uiMan.Add().ToPanel()
	header.Init(nil, ui.ElementTypePanel)
	header.DontFitContent()
	header.SetColor(shaderGraphNodeAccentColor)
	header.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	header.Base().Layout().SetZ(5.1)
	header.Base().Layout().Scale(shaderGraphNodeWidth, shaderGraphNodeHeaderH)
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
	n.title.Base().Layout().SetZ(5.2)
	n.title.Base().Layout().Scale(shaderGraphNodeWidth-shaderGraphNodePadding*2, shaderGraphNodeHeaderH)
	n.title.Base().Layout().SetOffset(shaderGraphNodePadding, 0)
	header.AddChild(n.title.Base())
}

func (n *shaderGraphNode) createDescription(uiMan *ui.Manager, description string) {
	n.description = uiMan.Add().ToLabel()
	n.description.Init(description)
	n.description.SetFontSize(9)
	n.description.SetColor(matrix.NewColor(0.70, 0.74, 0.80, 1))
	n.description.SetWidthAutoHeight(shaderGraphNodeWidth - shaderGraphNodePadding*2)
	n.description.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.description.Base().Layout().SetZ(5.1)
	n.description.Base().Layout().SetOffset(shaderGraphNodePadding, shaderGraphNodeHeaderH+6)
	n.root.AddChild(n.description.Base())
}

func (n *shaderGraphNode) createPorts(uiMan *ui.Manager, spec shaderGraphNodeSpec) {
	inputs := spec.Inputs
	outputs := spec.Outputs
	rowCount := max(len(inputs), len(outputs))
	startY := shaderGraphNodePortStartY + shaderGraphNodeFieldsHeight(spec)
	for i := range rowCount {
		y := startY + float32(i)*shaderGraphNodePortHeight
		if i < len(inputs) {
			n.inputs = append(n.inputs, n.createPort(uiMan, inputs[i], false, i, y))
		}
		if i < len(outputs) {
			n.outputs = append(n.outputs, n.createPort(uiMan, outputs[i], true, i, y))
		}
	}
}

func (n *shaderGraphNode) createPort(uiMan *ui.Manager, port shaderGraphPortSpec, output bool, index int, y float32) *shaderGraphPort {
	const dotSize = shaderGraphNodePortDotSize
	dotX := shaderGraphNodePadding
	labelX := shaderGraphNodePadding + dotSize + 5
	justify := rendering.FontJustifyLeft
	if output {
		dotX = shaderGraphNodeWidth - shaderGraphNodePadding - dotSize
		labelX = shaderGraphNodeWidth*0.5 - shaderGraphNodePadding
		justify = rendering.FontJustifyRight
	}

	dot := uiMan.Add().ToPanel()
	dot.Init(nil, ui.ElementTypePanel)
	dot.DontFitContent()
	dot.SetColor(shaderGraphPortColor(port.Type, output))
	dot.SetBorderRadius(dotSize*0.5, dotSize*0.5, dotSize*0.5, dotSize*0.5)
	dot.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	dot.Base().Layout().SetZ(5.2)
	dot.Base().Layout().Scale(dotSize, dotSize)
	dot.Base().Layout().SetOffset(dotX, y+6.5)
	n.bindSelectionEvent(dot.Base())
	n.root.AddChild(dot.Base())

	label := uiMan.Add().ToLabel()
	label.Init(shaderGraphPortLabel(port))
	label.SetFontSize(10)
	label.SetWrap(false)
	label.SetColor(matrix.NewColor(0.86, 0.88, 0.91, 1))
	label.SetJustify(justify)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(5.2)
	label.Base().Layout().Scale(shaderGraphNodeWidth*0.5-shaderGraphNodePadding*2-dotSize, shaderGraphNodePortHeight)
	label.Base().Layout().SetOffset(labelX, y)
	n.bindSelectionEvent(label.Base())
	n.root.AddChild(label.Base())

	graphPort := &shaderGraphPort{
		graph:       n.graph,
		node:        n,
		spec:        port,
		output:      output,
		index:       index,
		dot:         dot,
		label:       label,
		localAnchor: matrix.NewVec2(matrix.Float(dotX+dotSize*0.5), matrix.Float(y+6.5+dotSize*0.5)),
	}
	graphPort.bindEvents()
	return graphPort
}

func shaderGraphNodeHeight(spec shaderGraphNodeSpec) float32 {
	ports := max(len(spec.Inputs), len(spec.Outputs))
	return shaderGraphNodeBaseHeight + shaderGraphNodeFieldsHeight(spec) +
		float32(ports)*shaderGraphNodePortHeight
}

func shaderGraphNodeFieldsHeight(spec shaderGraphNodeSpec) float32 {
	if len(spec.Fields) == 0 {
		return 0
	}
	return float32(len(spec.Fields))*(shaderGraphFieldHeight+shaderGraphFieldGap) + 4
}

func shaderGraphPortLabel(port shaderGraphPortSpec) string {
	if port.Type == "" {
		return port.Name
	}
	return fmt.Sprintf("%s  %s", port.Name, port.Type)
}

func shaderGraphPortColor(portType string, output bool) matrix.Color {
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
		return shaderGraphSurfaceColor
	default:
		if output {
			return matrix.NewColor(0.76, 0.58, 0.92, 1)
		}
		return matrix.NewColor(0.58, 0.64, 0.72, 1)
	}
}
