package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/ui"
	"kaiju/uimarkup/css/helpers"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func setChildTextColor(elm markup.DocElement, color matrix.Color) {
	for _, c := range elm.HTML.Children {
		if c.DocumentElement.HTML.IsText() {
			c.DocumentElement.UI.(*ui.Label).SetColor(color)
		}
		setChildTextColor(*c.DocumentElement, color)
	}
}

func (p Color) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	} else {
		hex := values[0].Str
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
