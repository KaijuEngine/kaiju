/******************************************************************************/
/* css_grid_template_columns.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"strconv"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p GridTemplateColumns) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}

	if values[0].Str == "initial" || values[0].Str == "none" {
		panel.SetFlowLayout()
		panel.SetGridTemplateColumns(nil)
		return nil
	}

	// Explicit template, e.g. "8rem 1fr"
	if len(values) > 1 {
		cols := make([]float32, 0, len(values))
		for i := range values {
			s := strings.TrimSpace(values[i].Str)
			if strings.HasSuffix(s, "fr") {
				n := strings.TrimSpace(strings.TrimSuffix(s, "fr"))
				if n == "" {
					n = "1"
				}
				if f, err := strconv.ParseFloat(n, 32); err == nil && f > 0 {
					cols = append(cols, -float32(f))
					continue
				}
				cols = cols[:0]
				break
			}
			cols = append(cols, helpers.NumFromLength(s, host.Window))
		}
		if len(cols) == len(values) {
			panel.SetGridTemplateColumns(cols)
			return nil
		}
	}

	cols := 3
	str := values[0].Str
	if n, err := strconv.Atoi(str); err == nil && n > 0 {
		cols = n
	} else if strings.Contains(str, "repeat(") {
		if idx := strings.Index(str, "("); idx > 0 {
			part := strings.TrimSpace(str[idx+1:])
			if comma := strings.Index(part, ","); comma > 0 {
				if n, err := strconv.Atoi(strings.TrimSpace(part[:comma])); err == nil && n > 0 {
					cols = n
				}
			}
		}
	} else if len(values) > 1 {
		cols = len(values)
	}
	panel.SetGrid(cols)
	panel.SetGridTemplateColumns(nil)
	return nil
}
