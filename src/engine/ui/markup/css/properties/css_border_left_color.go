/******************************************************************************/
/* css_border_left_color.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

// color|transparent|initial|inherit
func (p BorderLeftColor) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderLeftColor requires 1 value")
	}

	value := values[0].Str
	switch value {
	case "transparent":
		colors := panel.BorderColor()
		panel.SetBorderColor(matrix.ColorTransparent(), colors[1], colors[2], colors[3])
		return nil
	case "initial":
		colors := panel.BorderColor()
		panel.SetBorderColor(matrix.ColorWhite(), colors[1], colors[2], colors[3])
		return nil
	case "inherit":
		if elm.Parent.Value() != nil {
			colors := elm.Parent.Value().UI.ToPanel().BorderColor()
			panel.SetBorderColor(colors[0], colors[1], colors[2], colors[3])
		}
		return nil
	}

	if newHex, ok := helpers.ColorMap[value]; ok {
		value = newHex
	}
	color, err := matrix.ColorFromHexString(value)
	if err != nil {
		return err
	}
	colors := panel.BorderColor()
	panel.SetBorderColor(color, colors[1], colors[2], colors[3])
	return nil
}
