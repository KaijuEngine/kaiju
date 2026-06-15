/******************************************************************************/
/* schema_node_inspector.go                                                   */
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
	schemaNodeInspectorW      = float32(320.0)
	schemaNodeInspectorHeader = float32(26.0)
	schemaNodeInspectorRowH   = float32(24.0)
	schemaNodeInspectorPad    = float32(8.0)
	schemaNodeInspectorGap    = float32(4.0)
)

func (g *schemaGraph) hidePropertyInspectorsExcept(active *schemaNode) {
	if g == nil {
		return
	}
	for i := range g.nodes {
		node := g.nodes[i]
		if node != nil && node != active && node.propertyInspector != nil {
			node.propertyInspector.Base().Hide()
		}
	}
}

func (n *schemaNode) addPropertyInspectorEvents(target *ui.UI) {
	if n == nil || n.kind != schemaNodeKindProperty || target == nil {
		return
	}
	target.AddEvent(ui.EventTypeRightDown, n.showPropertyInspector)
	target.AddEvent(ui.EventTypeRightClick, n.showPropertyInspector)
}

func (n *schemaNode) showPropertyInspector() {
	if n == nil || n.kind != schemaNodeKindProperty || n.graph == nil {
		return
	}
	n.graph.hidePropertyInspectorsExcept(n)
	n.ensurePropertyInspector()
	if n.propertyInspector != nil {
		n.propertyInspector.Base().Show()
	}
}

func (n *schemaNode) updatePropertyInspectorPosition() {
	if n == nil || n.propertyInspector == nil {
		return
	}
	n.propertyInspector.Base().Layout().SetOffset(
		float32(n.position.X())+n.width+10,
		float32(n.position.Y()))
}

func (n *schemaNode) ensurePropertyInspector() {
	if n.propertyInspector != nil || n.graph == nil {
		return
	}
	uiMan := &n.graph.uiMan
	fields := schemaNodePropertyTextFields()
	height := schemaNodeInspectorHeader + schemaNodeInspectorPad*2 +
		float32(len(fields))*schemaNodeInspectorRowH +
		float32(max(0, len(fields)-1))*schemaNodeInspectorGap

	panel := uiMan.Add().ToPanel()
	panel.Init(nil, ui.ElementTypePanel)
	panel.DontFitContent()
	panel.SetColor(schemaNodeBodyColor)
	panel.SetBorderRadius(4, 4, 4, 4)
	panel.SetBorderSize(1, 1, 1, 1)
	panel.SetBorderColor(schemaNodeAccentColor, schemaNodeAccentColor, schemaNodeAccentColor, schemaNodeAccentColor)
	panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	panel.Base().Layout().SetZ(4)
	panel.Base().Layout().Scale(schemaNodeInspectorW, height)
	panel.Base().AddEvent(ui.EventTypeMiss, func() {
		panel.Base().Hide()
	})
	panel.Base().Hide()
	n.graph.root.AddChild(panel.Base())
	n.propertyInspector = panel
	n.updatePropertyInspectorPosition()

	n.createPropertyInspectorHeader(uiMan, panel)
	y := schemaNodeInspectorHeader + schemaNodeInspectorPad
	for i := range fields {
		n.createPropertyInspectorRow(uiMan, panel, fields[i], y)
		y += schemaNodeInspectorRowH + schemaNodeInspectorGap
	}
}

func (n *schemaNode) createPropertyInspectorHeader(uiMan *ui.Manager, panel *ui.Panel) {
	header := uiMan.Add().ToPanel()
	header.Init(nil, ui.ElementTypePanel)
	header.DontFitContent()
	header.SetColor(schemaNodeAccentColor)
	header.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	header.Base().Layout().SetZ(0.1)
	header.Base().Layout().Scale(schemaNodeInspectorW, schemaNodeInspectorHeader)
	header.Base().Layout().SetOffset(0, 0)
	panel.AddChild(header.Base())

	title := uiMan.Add().ToLabel()
	title.Init("JSON Schema fields")
	title.SetFontSize(12)
	title.SetWrap(false)
	title.SetColor(matrix.ColorWhite())
	title.SetBaseline(rendering.FontBaselineCenter)
	title.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	title.Base().Layout().SetZ(0.2)
	title.Base().Layout().Scale(schemaNodeInspectorW-schemaNodeInspectorPad*2, schemaNodeInspectorHeader)
	title.Base().Layout().SetOffset(schemaNodeInspectorPad, schemaNodeTextOffsetY)
	header.AddChild(title.Base())
}

func (n *schemaNode) createPropertyInspectorRow(uiMan *ui.Manager, panel *ui.Panel, field schemaNodePropertyTextField, y float32) {
	row := uiMan.Add().ToPanel()
	row.Init(nil, ui.ElementTypePanel)
	row.DontFitContent()
	row.SetColor(schemaNodeRowColor)
	row.SetBorderRadius(2, 2, 2, 2)
	row.SetBorderSize(1, 1, 1, 1)
	row.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	row.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	row.Base().Layout().SetZ(0.1)
	row.Base().Layout().Scale(schemaNodeInspectorW-schemaNodeInspectorPad*2, schemaNodeInspectorRowH)
	row.Base().Layout().SetOffset(schemaNodeInspectorPad, y)
	panel.AddChild(row.Base())

	labelW := float32(88.0)
	label := uiMan.Add().ToLabel()
	label.Init(field.Label)
	label.SetFontSize(11)
	label.SetWrap(false)
	label.SetColor(schemaNodeLabelColor)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(0.2)
	label.Base().Layout().Scale(labelW, schemaNodeInspectorRowH)
	label.Base().Layout().SetOffset(8, schemaNodeTextOffsetY)
	row.AddChild(label.Base())

	input := uiMan.Add().ToInput()
	input.Init("")
	input.SetFontSize(11)
	input.SetFGColor(schemaNodeTextColor)
	input.SetBGColor(schemaNodeBodyColor)
	input.SetCursorColor(matrix.ColorWhite())
	input.SetSelectColor(schemaNodeAccentColor)
	input.SetTextWithoutEvent(n.propertyFieldValue(field.Key))
	input.SetType(ui.InputTypeText)

	inputPanel := input.Base().ToPanel()
	inputPanel.SetBorderRadius(2, 2, 2, 2)
	inputPanel.SetBorderSize(1, 1, 1, 1)
	inputPanel.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	input.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	input.Base().Layout().SetZ(0.3)
	input.Base().Layout().Scale(schemaNodeInspectorW-schemaNodeInspectorPad*2-labelW-12, schemaNodeInspectorRowH-4)
	input.Base().Layout().SetOffset(labelW+6, 2)
	input.Base().AddEvent(ui.EventTypeChange, func() {
		n.setPropertyField(field.Key, input.Text())
	})
	row.AddChild(input.Base())
}
