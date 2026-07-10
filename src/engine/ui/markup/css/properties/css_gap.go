/******************************************************************************/
/* css_gap.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (p Gap) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || values[0].Str == "initial" || values[0].Str == "inherit" {
		return nil
	}

	gapX := matrix.Float(8)
	gapY := matrix.Float(8)
	if len(values) > 0 {
		gapY = helpers.NumFromLength(values[0].Str, host.Window) // row-gap
		if len(values) > 1 {
			gapX = helpers.NumFromLength(values[1].Str, host.Window) // column-gap
		} else {
			gapX = gapY
		}
	}
	panel.SetGridGap(gapX, gapY)
	return nil
}
