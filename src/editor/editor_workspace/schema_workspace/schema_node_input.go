/******************************************************************************/
/* schema_node_input.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package schema_workspace

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/matrix"
)

func (n *schemaNode) createRowInput(uiMan *ui.Manager, rowPanel *ui.Panel, row schemaNodeRowSpec, x, width float32) {
	input := uiMan.Add().ToInput()
	input.Init("")
	input.SetFontSize(11)
	input.SetFGColor(schemaNodeTextColor)
	input.SetBGColor(schemaNodeBodyColor)
	input.SetCursorColor(matrix.ColorWhite())
	input.SetSelectColor(schemaNodeAccentColor)
	input.SetTextWithoutEvent(n.rowValue(row))
	input.SetType(ui.InputTypeText)

	panel := input.Base().ToPanel()
	panel.SetBorderRadius(2, 2, 2, 2)
	panel.SetBorderSize(1, 1, 1, 1)
	panel.SetBorderColor(schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor, schemaNodeBorderColor)
	input.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	input.Base().Layout().SetZ(0.3)
	input.Base().Layout().Scale(width-5, schemaNodeRowH-4)
	input.Base().Layout().SetOffset(x, 2)
	input.Base().AddEvent(ui.EventTypeChange, func() {
		if row.Kind == schemaNodeRowKindPropertyName {
			n.setPropertyName(input.Text())
		}
	})
	rowPanel.AddChild(input.Base())
}
