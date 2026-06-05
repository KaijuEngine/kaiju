/******************************************************************************/
/* shader_graph_node.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shading_workspace

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
	shaderGraphSurfaceColor    = matrix.NewColor(0.39, 0.82, 0.43, 1.0)
)

type shaderGraphNode struct {
	graph       *shaderGraph
	host        *engine.Host
	root        *ui.Panel
	bodyDrag    *ui.Panel
	title       *ui.Label
	description *ui.Label
	fields      []*shaderGraphNodeField
	inputs      []*shaderGraphPort
	outputs     []*shaderGraphPort
	values      map[string]shaderGraphNodeFieldValue
	position    matrix.Vec2
	dragging    bool
	dragMouse   matrix.Vec2
	dragOrigin  matrix.Vec2
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
	n.position = n.dragOrigin.Add(delta)
	n.applyViewOffset()
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
}

func (n *shaderGraphNode) stopDrag() {
	n.dragging = false
}

func (n *shaderGraphNode) applyViewOffset() {
	if n.root == nil {
		return
	}
	position := n.position
	if n.graph != nil {
		position = n.graph.viewPosition(position)
	}
	n.root.Base().Layout().SetOffset(position.X(), position.Y())
}

func (n *shaderGraphNode) bindDragEvents(target *ui.UI) {
	target.AddEvent(ui.EventTypeDown, n.beginDrag)
	target.AddEvent(ui.EventTypeUp, n.stopDrag)
	target.AddEvent(ui.EventTypeDragEnd, n.stopDrag)
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
	case "vec3":
		return matrix.NewColor(0.42, 0.76, 0.47, 1)
	case "color":
		return matrix.NewColor(0.91, 0.58, 0.30, 1)
	case "surface", "volume":
		return shaderGraphSurfaceColor
	default:
		if output {
			return matrix.NewColor(0.76, 0.58, 0.92, 1)
		}
		return matrix.NewColor(0.58, 0.64, 0.72, 1)
	}
}
