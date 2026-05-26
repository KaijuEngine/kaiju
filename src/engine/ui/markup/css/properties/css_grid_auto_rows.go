/******************************************************************************/
/* css_grid_auto_rows.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p GridAutoRows) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	size, err := gridAutoTrackSize(p.Key(), values, host)
	if err != nil {
		return err
	}
	panel.SetGridAutoRows(size)
	return nil
}
