/******************************************************************************/
/* css_border_top_style.go                                                    */
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
func (p BorderTopStyle) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderTopStyle requires 1 value")
	}
	border, ok := borderStyleFromStr(values[0].Str, 1, elm)
	if !ok {
		return errors.New("BorderTopStyle: invalid border style")
	}
	borders := panel.BorderStyle()
	panel.SetBorderStyle(borders[0], border, borders[2], borders[3])
	return nil
}
