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

func setChildTextBackgroundColor(elm document.DocElement, color matrix.Color) {
	for _, c := range elm.HTML.Children {
		if c.DocumentElement.HTML.IsText() {
			c.DocumentElement.UI.(*ui.Label).SetBGColor(color)
		}
		setChildTextBackgroundColor(*c.DocumentElement, color)
	}
}

func (p BackgroundColor) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	}
	var err error
	var color matrix.Color
	hex := values[0].Str
	if hex == "inherit" {
		panel.OnRender.Add(func() {
			if panel.Entity().Parent != nil {
				p := ui.FirstPanelOnEntity(panel.Entity().Parent)
				panel.SetColor(p.ShaderData().FgColor)
			}
		})
		return nil
	} else {
		if newHex, ok := helpers.ColorMap[hex]; ok {
			hex = newHex
		}
		if color, err = matrix.ColorFromHexString(hex); err == nil {
			panel.SetColor(color)
			setChildTextBackgroundColor(elm, color)
			return nil
		} else {
			return err
		}
	}
}
