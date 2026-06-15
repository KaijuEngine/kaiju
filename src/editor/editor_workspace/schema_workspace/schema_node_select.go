/******************************************************************************/
/* schema_node_select.go                                                      */
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
	schemaNodeSelectOptionH = float32(22.0)
	schemaNodeSelectArrowW  = float32(16.0)
)

type schemaNodeFloatingPanel struct {
	panel  *ui.Panel
	offset matrix.Vec2
}

func (n *schemaNode) addFloatingPanel(panel *ui.Panel, x, y float32) {
	if n == nil || panel == nil {
		return
	}
	n.floatingPanels = append(n.floatingPanels, schemaNodeFloatingPanel{
		panel:  panel,
		offset: matrix.NewVec2(matrix.Float(x), matrix.Float(y)),
	})
	n.updateFloatingPanels()
}

func (n *schemaNode) updateFloatingPanels() {
	if n == nil {
		return
	}
	for i := range n.floatingPanels {
		floating := &n.floatingPanels[i]
		if floating.panel == nil {
			continue
		}
		floating.panel.Base().Layout().SetOffset(
			float32(n.position.X()+floating.offset.X()),
			float32(n.position.Y()+floating.offset.Y()))
	}
}

func (n *schemaNode) createRowSelect(uiMan *ui.Manager, rowPanel *ui.Panel, row schemaNodeRowSpec, x, width, y float32) {
	controlW := width - 5
	controlH := schemaNodeRowH - 4
	control := uiMan.Add().ToPanel()
	control.Init(nil, ui.ElementTypePanel)
	control.DontFitContent()
	control.SetColor(schemaNodeBodyColor)
	control.SetBorderRadius(2, 2, 2, 2)
	control.SetBorderSize(1, 1, 1, 1)
	control.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	control.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	control.Base().Layout().SetZ(0.3)
	control.Base().Layout().Scale(controlW, controlH)
	control.Base().Layout().SetOffset(x, 2)
	rowPanel.AddChild(control.Base())

	label := uiMan.Add().ToLabel()
	label.Init(n.rowValue(row))
	label.SetFontSize(11)
	label.SetWrap(false)
	label.SetColor(schemaNodeTextColor)
	label.SetBaseline(rendering.FontBaselineCenter)
	label.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	label.Base().Layout().SetZ(0.1)
	label.Base().Layout().Scale(controlW-schemaNodeSelectArrowW-7, controlH)
	label.Base().Layout().SetOffset(5, schemaNodeTextOffsetY)
	control.AddChild(label.Base())

	arrow := uiMan.Add().ToLabel()
	arrow.Init("v")
	arrow.SetFontSize(12)
	arrow.SetWrap(false)
	arrow.SetColor(schemaNodeTextColor)
	arrow.SetJustify(rendering.FontJustifyCenter)
	arrow.SetBaseline(rendering.FontBaselineCenter)
	arrow.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	arrow.Base().Layout().SetZ(0.1)
	arrow.Base().Layout().Scale(schemaNodeSelectArrowW, controlH)
	arrow.Base().Layout().SetOffset(controlW-schemaNodeSelectArrowW, schemaNodeTextOffsetY)
	control.AddChild(arrow.Base())

	list := n.createRowSelectList(uiMan, row, y+controlH+4, x, controlW, label)
	control.Base().AddEvent(ui.EventTypeClick, func() {
		if list.Base().IsActive() {
			list.Base().Hide()
		} else {
			list.Base().Show()
		}
	})
	control.Base().AddEvent(ui.EventTypeEnter, func() {
		control.SetColor(schemaNodeBorderColor)
	})
	control.Base().AddEvent(ui.EventTypeExit, func() {
		control.SetColor(schemaNodeBodyColor)
	})
}

func (n *schemaNode) createRowSelectList(uiMan *ui.Manager, row schemaNodeRowSpec, y, x, width float32, label *ui.Label) *ui.Panel {
	options := schemaNodeTypeOptions()
	list := uiMan.Add().ToPanel()
	list.Init(nil, ui.ElementTypePanel)
	list.DontFitContent()
	list.SetColor(schemaNodeBodyColor)
	list.SetBorderRadius(2, 2, 2, 2)
	list.SetBorderSize(1, 1, 1, 1)
	list.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	list.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	list.Base().Layout().SetZ(0.8)
	list.Base().Layout().Scale(width, schemaNodeSelectOptionH*float32(len(options)))
	list.Base().Layout().SetOffset(schemaNodePadding+x, y)
	list.Base().AddEvent(ui.EventTypeMiss, func() {
		list.Base().Hide()
	})
	list.Base().Hide()
	if n.graph != nil && n.graph.root != nil {
		n.graph.root.AddChild(list.Base())
		n.addFloatingPanel(list, schemaNodePadding+x, y)
	} else {
		list.Base().Layout().SetOffset(schemaNodePadding+x, y)
		n.root.AddChild(list.Base())
	}

	for i, option := range options {
		n.createRowSelectOption(uiMan, list, row, option, float32(i)*schemaNodeSelectOptionH, width, label)
	}
	return list
}

func (n *schemaNode) createRowSelectOption(uiMan *ui.Manager, list *ui.Panel, row schemaNodeRowSpec, option string, y, width float32, label *ui.Label) {
	item := uiMan.Add().ToPanel()
	item.Init(nil, ui.ElementTypePanel)
	item.DontFitContent()
	item.SetColor(schemaNodeBodyColor)
	item.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	item.Base().Layout().SetZ(0.1)
	item.Base().Layout().Scale(width, schemaNodeSelectOptionH)
	item.Base().Layout().SetOffset(0, y)
	item.Base().AddEvent(ui.EventTypeClick, func() {
		n.pickRowSelectOption(row, option, label)
		list.Base().Hide()
	})
	item.Base().AddEvent(ui.EventTypeEnter, func() {
		item.SetColor(schemaNodeBorderColor)
	})
	item.Base().AddEvent(ui.EventTypeExit, func() {
		item.SetColor(schemaNodeBodyColor)
	})
	list.AddChild(item.Base())

	itemLabel := uiMan.Add().ToLabel()
	itemLabel.Init(option)
	itemLabel.SetFontSize(11)
	itemLabel.SetWrap(false)
	itemLabel.SetColor(schemaNodeTextColor)
	itemLabel.SetBaseline(rendering.FontBaselineCenter)
	itemLabel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	itemLabel.Base().Layout().SetZ(0.1)
	itemLabel.Base().Layout().Scale(width-10, schemaNodeSelectOptionH)
	itemLabel.Base().Layout().SetOffset(5, schemaNodeTextOffsetY)
	item.AddChild(itemLabel.Base())
}

func (n *schemaNode) pickRowSelectOption(row schemaNodeRowSpec, option string, label *ui.Label) {
	if row.Kind == schemaNodeRowKindSchemaType {
		n.setSchemaType(option)
		label.SetText(n.rowValue(row))
	}
}
