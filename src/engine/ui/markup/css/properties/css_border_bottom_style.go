/******************************************************************************/
/* css_border_bottom_style.go                                                 */
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
func (p BorderBottomStyle) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderBottomStyle requires 1 value")
	}
	border, ok := borderStyleFromStr(values[0].Str, 3, elm)
	if !ok {
		return errors.New("BorderBottomStyle: invalid border style")
	}
	borders := panel.BorderStyle()
	panel.SetBorderStyle(borders[0], borders[1], borders[2], border)
	return nil
}
