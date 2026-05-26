/******************************************************************************/
/* css_border_left_style.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

// none|hidden|dotted|dashed|solid|double|groove|ridge|inset|outset|initial|inherit
func (p BorderLeftStyle) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderLeftStyle requires 1 value")
	}

	border, ok := borderStyleFromStr(values[0].Str, 0, elm)
	if !ok {
		return errors.New("BorderLeftStyle: invalid border style")
	}

	borders := panel.BorderStyle()
	panel.SetBorderStyle(border, borders[1], borders[2], borders[3])
	return nil
}
