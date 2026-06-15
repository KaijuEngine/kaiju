/******************************************************************************/
/* schema_node_checkbox.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func (n *schemaNode) createRowCheckbox(uiMan *ui.Manager, rowPanel *ui.Panel, row schemaNodeRowSpec, x float32) {
	box := uiMan.Add().ToPanel()
	box.Init(nil, ui.ElementTypePanel)
	box.DontFitContent()
	box.SetColor(schemaNodeBodyColor)
	box.SetBorderRadius(2, 2, 2, 2)
	box.SetBorderSize(1, 1, 1, 1)
	box.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	box.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	box.Base().Layout().SetZ(0.3)
	box.Base().Layout().Scale(18, 18)
	box.Base().Layout().SetOffset(x+2, 3)
	rowPanel.AddChild(box.Base())

	check := uiMan.Add().ToLabel()
	check.Init("x")
	check.SetFontSize(12)
	check.SetWrap(false)
	check.SetColor(matrix.ColorWhite())
	check.SetJustify(rendering.FontJustifyCenter)
	check.SetBaseline(rendering.FontBaselineCenter)
	check.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	check.Base().Layout().SetZ(0.1)
	check.Base().Layout().Scale(18, 18)
	check.Base().Layout().SetOffset(0, schemaNodeTextOffsetY)
	box.AddChild(check.Base())

	if row.Kind == schemaNodeRowKindRequired {
		check.Base().SetVisibility(n.propertyRequired)
		toggleRequired := func() {
			n.setPropertyRequired(!n.propertyRequired)
			check.Base().SetVisibility(n.propertyRequired)
		}
		box.Base().AddEvent(ui.EventTypeClick, toggleRequired)
		box.Base().AddEvent(ui.EventTypeEnter, func() {
			box.SetColor(schemaNodeBorderColor)
		})
		box.Base().AddEvent(ui.EventTypeExit, func() {
			box.SetColor(schemaNodeBodyColor)
		})
	}
}
