/******************************************************************************/
/* css_color.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func setChildTextColor(elm *document.Element, color matrix.Color) {
	for _, c := range elm.Children {
		if c.IsText() {
			c.UI.ToLabel().SetColor(color)
		}
		setChildTextColor(c, color)
	}
}

func (p Color) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	hex := values[0].Str
	if hex == "inherit" {
		// TODO:  If a text color is set on a parent somewhere, that parent may not have any text in it
		// so we'll need to store the color on the panel or something?
		return nil
	}

	if newHex, ok := helpers.ColorMap[hex]; ok {
		hex = newHex
	}

	color, err := matrix.ColorFromHexString(hex)
	if err != nil {
		return err
	}

	if panel.Base().IsType(ui.ElementTypeInput) {
		panel.Base().ToInput().SetFGColor(color)
		return nil
	}
	if panel.Base().IsType(ui.ElementTypeTextArea) {
		panel.Base().ToTextArea().SetFGColor(color)
		return nil
	}

	setChildTextColor(elm, color)
	return nil
}
