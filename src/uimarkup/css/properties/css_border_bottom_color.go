package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// color|transparent|initial|inherit
func (p BorderBottomColor) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderLeftColor requires 1 value")
	} else {
		if values[0].Str == "transparent" {
			colors := panel.BorderColor()
			panel.SetBorderColor(colors[0], colors[1], colors[2], matrix.ColorTransparent())
			return nil
		} else if values[0].Str == "initial" {
			colors := panel.BorderColor()
			panel.SetBorderColor(colors[0], colors[1], colors[2], matrix.ColorWhite())
			return nil
		} else if values[0].Str == "inherit" {
			if elm.HTML.Parent != nil {
				colors := elm.HTML.Parent.DocumentElement.UI.(*ui.Panel).BorderColor()
				panel.SetBorderColor(colors[0], colors[1], colors[2], colors[3])
			}
			return nil
		} else if color, err := matrix.ColorFromHexString(values[0].Str); err != nil {
			return err
		} else {
			colors := panel.BorderColor()
			panel.SetBorderColor(colors[0], colors[1], colors[2], color)
			return nil
		}
	}
}
