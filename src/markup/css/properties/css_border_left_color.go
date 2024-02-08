package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
)

// color|transparent|initial|inherit
func (p BorderLeftColor) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderLeftColor requires 1 value")
	} else {
		if values[0].Str == "transparent" {
			colors := panel.BorderColor()
			panel.SetBorderColor(matrix.ColorTransparent(), colors[1], colors[2], colors[3])
			return nil
		} else if values[0].Str == "initial" {
			colors := panel.BorderColor()
			panel.SetBorderColor(matrix.ColorWhite(), colors[1], colors[2], colors[3])
			return nil
		} else if values[0].Str == "inherit" {
			if elm.HTML.Parent != nil {
				colors := elm.HTML.Parent.DocumentElement.UI.(*ui.Panel).BorderColor()
				panel.SetBorderColor(colors[0], colors[1], colors[2], colors[3])
			}
			return nil
		} else {
			hex := values[0].Str
			if newHex, ok := helpers.ColorMap[values[0].Str]; ok {
				hex = newHex
			}
			if color, err := matrix.ColorFromHexString(hex); err != nil {
				return err
			} else {
				colors := panel.BorderColor()
				panel.SetBorderColor(color, colors[1], colors[2], colors[3])
				return nil
			}
		}
	}
}
