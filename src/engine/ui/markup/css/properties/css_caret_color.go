/******************************************************************************/
/* css_caret_color.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/functions"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (p CaretColor) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("CaretColor requires exactly 1 value")
	}

	hex := values[0].Str
	if hex == "auto" || hex == "inherit" {
		return nil
	}
	switch hex {
	case "rgb":
		hex, _ = functions.Rgb{}.Process(panel, elm, values[0])
	case "rgba":
		hex, _ = functions.Rgba{}.Process(panel, elm, values[0])
	}
	if newHex, ok := helpers.ColorMap[hex]; ok {
		hex = newHex
	}
	color, err := matrix.ColorFromHexString(hex)
	if err != nil {
		return err
	}
	switch panel.Base().Type() {
	case ui.ElementTypeInput:
		panel.Base().ToInput().SetCursorColor(color)
	case ui.ElementTypeTextArea:
		panel.Base().ToTextArea().SetCursorColor(color)
	}
	return nil
}
