/******************************************************************************/
/* schema_node.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

const (
	schemaNodeWidth       = float32(260.0)
	schemaNodeHeaderH     = float32(28.0)
	schemaNodePadding     = float32(10.0)
	schemaNodeSummaryH    = float32(20.0)
	schemaNodeContentGap  = float32(5.0)
	schemaNodeRowH        = float32(24.0)
	schemaNodeRowGap      = float32(4.0)
	schemaNodeActionH     = float32(24.0)
	schemaNodeActionGap   = float32(5.0)
	schemaNodeBottomPad   = float32(8.0)
	schemaNodeBorderWidth = float32(1.0)
	schemaNodeTextOffsetY = float32(2.0)
)

var (
	schemaNodeAccentColor    = matrix.NewColor(0.4078, 0.1647, 0.1765, 1)
	schemaNodeBodyColor      = matrix.NewColor(0.12, 0.13, 0.15, 1)
	schemaNodeRowColor       = matrix.NewColor(0.085, 0.095, 0.115, 1)
	schemaNodeBorderColor    = matrix.NewColor(0.22, 0.24, 0.29, 1)
	schemaNodeLabelColor     = matrix.NewColor(0.74, 0.78, 0.84, 1)
	schemaNodeTextColor      = matrix.NewColor(0.88, 0.90, 0.94, 1)
	schemaNodeMutedTextColor = matrix.NewColor(0.70, 0.74, 0.80, 1)
)

type schemaNode struct {
	graph    *schemaGraph
	root     *ui.Panel
	id       string
	kind     schemaNodeKind
	parent   *schemaNode
	children []*schemaNode
	position matrix.Vec2
	width    float32
	height   float32

	propertyName     string
	definitionName   string
	schemaType       string
	propertyRequired bool
	propertyFields   map[string]string

	titleLabel        *ui.Label
	requiredMarker    *ui.Label
	propertyInspector *ui.Panel
	floatingPanels    []schemaNodeFloatingPanel
}

func (n *schemaNode) Initialize(uiMan *ui.Manager, parent *ui.Panel, spec schemaNodeSpec) {
	n.width = max(schemaNodeWidth, spec.MinWidth)
	n.height = schemaNodeHeight(spec)
	n.root = uiMan.Add().ToPanel()
	n.root.Init(nil, ui.ElementTypePanel)
	n.root.DontFitContent()
	n.root.SetColor(schemaNodeBodyColor)
	n.root.SetBorderRadius(6, 6, 6, 6)
	n.root.SetBorderSize(schemaNodeBorderWidth, schemaNodeBorderWidth, schemaNodeBorderWidth, schemaNodeBorderWidth)
	n.root.SetBorderColor(spec.Accent, spec.Accent, spec.Accent, spec.Accent)
	n.root.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	n.root.Base().Layout().SetZ(0.2)
	n.root.Base().Layout().Scale(n.width, n.height)
	parent.AddChild(n.root.Base())

	n.createHeader(uiMan, spec)
	n.createSummary(uiMan, spec)
	nextY := n.createRows(uiMan, spec)
	n.createActions(uiMan, spec, nextY)
}

func (n *schemaNode) SetPosition(position matrix.Vec2) {
	n.position = position
	if n.root == nil {
		return
	}
	n.root.Base().Layout().SetOffset(float32(position.X()), float32(position.Y()))
	n.updatePropertyInspectorPosition()
	n.updateFloatingPanels()
}

func (n *schemaNode) createHeader(uiMan *ui.Manager, spec schemaNodeSpec) {
	header := uiMan.Add().ToPanel()
	header.Init(nil, ui.ElementTypePanel)
	header.DontFitContent()
	header.SetColor(spec.Accent)
	header.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	header.Base().Layout().SetZ(0.1)
	header.Base().Layout().Scale(n.width, schemaNodeHeaderH)
	header.Base().Layout().SetOffset(0, 0)
	n.root.AddChild(header.Base())
	n.addPropertyInspectorEvents(header.Base())

	title := uiMan.Add().ToLabel()
	title.Init(spec.Title)
	title.SetFontSize(13)
	title.SetWrap(false)
	title.SetColor(matrix.ColorWhite())
	title.SetBaseline(rendering.FontBaselineCenter)
	title.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	title.Base().Layout().SetZ(0.2)
	title.Base().Layout().Scale(n.width-schemaNodePadding*2-24, schemaNodeHeaderH)
	title.Base().Layout().SetOffset(schemaNodePadding, schemaNodeTextOffsetY)
	header.AddChild(title.Base())
	n.titleLabel = title
	n.addPropertyInspectorEvents(title.Base())

	if spec.Kind == schemaNodeKindProperty {
		marker := uiMan.Add().ToLabel()
		marker.Init("*")
		marker.SetFontSize(16)
		marker.SetWrap(false)
		marker.SetColor(matrix.ColorWhite())
		marker.SetJustify(rendering.FontJustifyRight)
		marker.SetBaseline(rendering.FontBaselineCenter)
		marker.Base().Layout().SetPositioning(ui.PositioningAbsolute)
		marker.Base().Layout().SetZ(0.5)
		marker.Base().Layout().Scale(24, schemaNodeHeaderH)
		marker.Base().Layout().SetOffset(n.width-schemaNodePadding-24, schemaNodeTextOffsetY)
		header.AddChild(marker.Base())
		n.requiredMarker = marker
		n.refreshRequiredMarker()
		n.addPropertyInspectorEvents(n.root.Base())
		n.addPropertyInspectorEvents(marker.Base())
	}
}

func (n *schemaNode) createSummary(uiMan *ui.Manager, spec schemaNodeSpec) {
	summary := uiMan.Add().ToLabel()
	summary.Init(spec.Summary)
	summary.SetFontSize(11)
	summary.SetWrap(false)
	summary.SetColor(schemaNodeMutedTextColor)
	summary.SetBaseline(rendering.FontBaselineCenter)
	summary.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	summary.Base().Layout().SetZ(0.1)
	summary.Base().Layout().Scale(n.width-schemaNodePadding*2, schemaNodeSummaryH)
	summary.Base().Layout().SetOffset(schemaNodePadding,
		schemaNodeHeaderH+schemaNodeContentGap+schemaNodeTextOffsetY)
	n.root.AddChild(summary.Base())
}

func (n *schemaNode) createRows(uiMan *ui.Manager, spec schemaNodeSpec) float32 {
	y := schemaNodeBodyStartY()
	for i := range spec.Rows {
		n.createRow(uiMan, spec.Rows[i], y)
		y += schemaNodeRowH + schemaNodeRowGap
	}
	if len(spec.Rows) > 0 {
		y -= schemaNodeRowGap
		y += schemaNodeActionGap
	}
	return y
}

func (n *schemaNode) createRow(uiMan *ui.Manager, row schemaNodeRowSpec, y float32) {
	rowPanel := uiMan.Add().ToPanel()
	rowPanel.Init(nil, ui.ElementTypePanel)
	rowPanel.DontFitContent()
	rowPanel.SetColor(schemaNodeRowColor)
	rowPanel.SetBorderRadius(3, 3, 3, 3)
	rowPanel.SetBorderSize(1, 1, 1, 1)
	rowPanel.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	rowPanel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	rowPanel.Base().Layout().SetZ(0.1)
	rowPanel.Base().Layout().Scale(n.width-schemaNodePadding*2, schemaNodeRowH)
	rowPanel.Base().Layout().SetOffset(schemaNodePadding, y)
	n.root.AddChild(rowPanel.Base())
	n.addPropertyInspectorEvents(rowPanel.Base())

	labelWidth := (n.width - schemaNodePadding*2) * 0.48
	valueWidth := (n.width - schemaNodePadding*2) - labelWidth
	label := uiMan.Add().ToLabel()
	label.Init(row.Label)
	label.SetFontSize(11)
	label.SetWrap(false)
	label.SetColor(schemaNodeLabelColor)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(0.2)
	label.Base().Layout().Scale(labelWidth-schemaNodePadding, schemaNodeRowH)
	label.Base().Layout().SetOffset(8, schemaNodeTextOffsetY)
	rowPanel.AddChild(label.Base())
	n.addPropertyInspectorEvents(label.Base())

	if schemaNodeRowIsEditable(row) {
		n.createRowInput(uiMan, rowPanel, row, labelWidth, valueWidth)
		return
	}
	if schemaNodeRowIsSelectable(row) {
		n.createRowSelect(uiMan, rowPanel, row, labelWidth, valueWidth, y)
		return
	}
	if schemaNodeRowIsCheckable(row) {
		n.createRowCheckbox(uiMan, rowPanel, row, labelWidth)
		return
	}

	value := uiMan.Add().ToLabel()
	value.Init(n.rowValue(row))
	value.SetFontSize(11)
	value.SetWrap(false)
	value.SetColor(schemaNodeTextColor)
	value.SetBaseline(rendering.FontBaselineCenter)
	value.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	value.Base().Layout().SetZ(0.2)
	value.Base().Layout().Scale(valueWidth-8, schemaNodeRowH)
	value.Base().Layout().SetOffset(labelWidth, schemaNodeTextOffsetY)
	rowPanel.AddChild(value.Base())
}

func (n *schemaNode) createActions(uiMan *ui.Manager, spec schemaNodeSpec, y float32) {
	for i := range spec.Actions {
		n.createAction(uiMan, spec.Actions[i], y)
		y += schemaNodeActionH + schemaNodeActionGap
	}
}

func (n *schemaNode) createAction(uiMan *ui.Manager, action schemaNodeActionSpec, y float32) {
	button := uiMan.Add().ToPanel()
	button.Init(nil, ui.ElementTypePanel)
	button.DontFitContent()
	button.SetColor(schemaNodeRowColor)
	button.SetBorderRadius(3, 3, 3, 3)
	button.SetBorderSize(1, 1, 1, 1)
	button.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	button.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	button.Base().Layout().SetZ(0.2)
	button.Base().Layout().Scale(n.width-schemaNodePadding*2, schemaNodeActionH)
	button.Base().Layout().SetOffset(schemaNodePadding, y)
	button.Base().AddEvent(ui.EventTypeClick, func() {
		n.executeAction(action.Kind)
	})
	button.Base().AddEvent(ui.EventTypeEnter, func() {
		button.SetColor(schemaNodeBorderColor)
	})
	button.Base().AddEvent(ui.EventTypeExit, func() {
		button.SetColor(schemaNodeRowColor)
	})
	n.root.AddChild(button.Base())

	label := uiMan.Add().ToLabel()
	label.Init(action.Label)
	label.SetFontSize(11)
	label.SetWrap(false)
	label.SetColor(schemaNodeTextColor)
	label.SetJustify(rendering.FontJustifyCenter)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(0.3)
	label.Base().Layout().Scale(n.width-schemaNodePadding*2, schemaNodeActionH)
	label.Base().Layout().SetOffset(0, schemaNodeTextOffsetY)
	button.AddChild(label.Base())
}

func (n *schemaNode) executeAction(kind schemaNodeActionKind) {
	switch kind {
	case schemaNodeActionAddProperties:
		if n.graph != nil {
			n.graph.AddProperties(n)
		}
	case schemaNodeActionAddProperty:
		if n.graph != nil {
			n.graph.AddProperty(n)
		}
	case schemaNodeActionAddDefinition:
		if n.graph != nil {
			n.graph.AddDefinition(n)
		}
	}
}

func schemaNodeHeight(spec schemaNodeSpec) float32 {
	rowCount := len(spec.Rows)
	contentBottom := schemaNodeBodyStartY()
	if rowCount > 0 {
		contentBottom += float32(rowCount)*schemaNodeRowH +
			float32(rowCount-1)*schemaNodeRowGap
	}
	actionCount := len(spec.Actions)
	if actionCount > 0 {
		if rowCount > 0 {
			contentBottom += schemaNodeActionGap
		}
		contentBottom += float32(actionCount)*schemaNodeActionH +
			float32(actionCount-1)*schemaNodeActionGap
	}
	return contentBottom + schemaNodeBottomPad
}

func schemaNodeBodyStartY() float32 {
	return schemaNodeHeaderH + schemaNodeContentGap + schemaNodeSummaryH + schemaNodeContentGap
}
