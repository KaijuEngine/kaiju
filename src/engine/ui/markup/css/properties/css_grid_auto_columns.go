/******************************************************************************/
/* css_grid_auto_columns.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p GridAutoColumns) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	size, err := gridAutoTrackSize(p.Key(), values, host)
	if err != nil {
		return err
	}
	panel.SetGridAutoColumns(size)
	return nil
}

func gridAutoTrackSize(property string, values []rules.PropertyValue, host *engine.Host) (float32, error) {
	if len(values) == 0 {
		return 0, nil
	}
	if len(values) != 1 {
		return 0, fmt.Errorf("%s expects exactly one track size", property)
	}
	switch values[0].Str {
	case "auto", "initial", "inherit", "unset":
		return 0, nil
	}
	size := helpers.NumFromLength(values[0].Str, host.Window)
	if size <= 0 {
		return 0, fmt.Errorf("%s expects a positive length", property)
	}
	return size, nil
}
