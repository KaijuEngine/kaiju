package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
)

func setChildTextColor(elm document.DocElement, color matrix.Color) {
	for _, c := range elm.HTML.Children {
		if c.DocumentElement.HTML.IsText() {
			c.DocumentElement.UI.(*ui.Label).SetColor(color)
		}
		setChildTextColor(*c.DocumentElement, color)
	}
}

func (p Color) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	} else {
		hex := values[0].Str
		if hex == "inherit" {
			// TODO:  If a text color is set on a parent somewhere, that parent may not have any text in it
			// so we'll need to store the color on the panel or something?
			return nil
		} else {
			if newHex, ok := helpers.ColorMap[hex]; ok {
				hex = newHex
			}
			if color, err := matrix.ColorFromHexString(hex); err == nil {
				setChildTextColor(elm, color)
				return nil
			} else {
				return err
			}
		}
	}
}
