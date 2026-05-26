/******************************************************************************/
/* css_grid.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"strconv"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Grid) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	if values[0].Str == "initial" || values[0].Str == "none" {
		return nil
	}
	// Shorthand can specify template columns/rows; for now enable grid with parsed or default columns
	cols := 3
	if len(values) > 0 {
		str := values[0].Str
		if n, err := strconv.Atoi(str); err == nil && n > 0 {
			cols = n
		} else if strings.Contains(str, "repeat(") {
			// crude parse for repeat(N, ...)
			if idx := strings.Index(str, "("); idx > 0 {
				part := strings.TrimSpace(str[idx+1:])
				if comma := strings.Index(part, ","); comma > 0 {
					if n, err := strconv.Atoi(strings.TrimSpace(part[:comma])); err == nil && n > 0 {
						cols = n
					}
				}
			}
		} else if len(values) > 1 {
			// multiple values may indicate columns
			cols = len(values)
		}
	}
	panel.SetGrid(cols)
	return nil
}
