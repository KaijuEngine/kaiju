/******************************************************************************/
/* css_border_left_width.go                                                   */
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

// medium|thin|thick|length|initial|inherit
func (p BorderLeftWidth) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderTopWidth requires exactly 1 value")
	}

	current := panel.Base().Layout().Border()
	size := borderSizeFromStr(values[0].Str, host.Window, current.X())
	panel.SetBorderSize(size, current.Y(), current.Z(), current.W())
	return nil
}
