/******************************************************************************/
/* css_row_gap.go                                                             */
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
)

func (p RowGap) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || values[0].Str == "initial" || values[0].Str == "inherit" {
		return nil
	}

	gap := helpers.NumFromLength(values[0].Str, host.Window)
	current := panel.GridGap()
	panel.SetGridGap(current.X(), gap)
	return nil
}
