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
	schemaNodePadding     = float32(12.0)
	schemaNodeSummaryH    = float32(18.0)
	schemaNodeRowH        = float32(24.0)
	schemaNodeRowGap      = float32(6.0)
	schemaNodeBottomPad   = float32(12.0)
	schemaNodeBorderWidth = float32(1.0)
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
	n.createRows(uiMan, spec)
}

func (n *schemaNode) SetPosition(position matrix.Vec2) {
	n.position = position
	if n.root == nil {
		return
	}
	n.root.Base().Layout().SetOffset(float32(position.X()), float32(position.Y()))
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

	title := uiMan.Add().ToLabel()
	title.Init(spec.Title)
	title.SetFontSize(12)
	title.SetWrap(false)
	title.SetColor(matrix.ColorWhite())
	title.SetBaseline(rendering.FontBaselineCenter)
	title.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	title.Base().Layout().SetZ(0.2)
	title.Base().Layout().Scale(n.width-schemaNodePadding*2, schemaNodeHeaderH)
	title.Base().Layout().SetOffset(schemaNodePadding, 0)
	header.AddChild(title.Base())
}

func (n *schemaNode) createSummary(uiMan *ui.Manager, spec schemaNodeSpec) {
	summary := uiMan.Add().ToLabel()
	summary.Init(spec.Summary)
	summary.SetFontSize(10)
	summary.SetWrap(false)
	summary.SetColor(schemaNodeMutedTextColor)
	summary.SetBaseline(rendering.FontBaselineCenter)
	summary.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	summary.Base().Layout().SetZ(0.1)
	summary.Base().Layout().Scale(n.width-schemaNodePadding*2, schemaNodeSummaryH)
	summary.Base().Layout().SetOffset(schemaNodePadding, schemaNodeHeaderH+7)
	n.root.AddChild(summary.Base())
}

func (n *schemaNode) createRows(uiMan *ui.Manager, spec schemaNodeSpec) {
	y := schemaNodeHeaderH + 31
	for i := range spec.Rows {
		n.createRow(uiMan, spec.Rows[i], y)
		y += schemaNodeRowH + schemaNodeRowGap
	}
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

	labelWidth := (n.width - schemaNodePadding*2) * 0.48
	valueWidth := (n.width - schemaNodePadding*2) - labelWidth
	label := uiMan.Add().ToLabel()
	label.Init(row.Label)
	label.SetFontSize(10)
	label.SetWrap(false)
	label.SetColor(schemaNodeLabelColor)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(0.2)
	label.Base().Layout().Scale(labelWidth-schemaNodePadding, schemaNodeRowH)
	label.Base().Layout().SetOffset(8, 0)
	rowPanel.AddChild(label.Base())

	value := uiMan.Add().ToLabel()
	value.Init(row.Value)
	value.SetFontSize(10)
	value.SetWrap(false)
	value.SetColor(schemaNodeTextColor)
	value.SetBaseline(rendering.FontBaselineCenter)
	value.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	value.Base().Layout().SetZ(0.2)
	value.Base().Layout().Scale(valueWidth-8, schemaNodeRowH)
	value.Base().Layout().SetOffset(labelWidth, 0)
	rowPanel.AddChild(value.Base())
}

func schemaNodeHeight(spec schemaNodeSpec) float32 {
	rowCount := len(spec.Rows)
	rowHeight := float32(0)
	if rowCount > 0 {
		rowHeight = float32(rowCount)*schemaNodeRowH + float32(rowCount-1)*schemaNodeRowGap
	}
	return schemaNodeHeaderH + 31 + rowHeight + schemaNodeBottomPad
}
