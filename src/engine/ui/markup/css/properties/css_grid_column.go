/******************************************************************************/
/* css_grid_column.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p GridColumn) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	startValues, endValues := splitGridLineValues(values)
	start, err := parseGridLineValue(startValues, p.Key())
	if err != nil {
		return err
	}
	end, err := parseGridLineValue(endValues, p.Key())
	if err != nil {
		return err
	}
	startLine := start.line
	endLine := end.line
	if start.isSpan {
		if endLine == 0 {
			return fmt.Errorf("grid-column start span requires an explicit end line")
		}
		startLine = endLine - start.span
		if startLine < 1 {
			startLine = 1
		}
	}
	if end.isSpan {
		if startLine == 0 {
			return fmt.Errorf("grid-column end span requires an explicit start line")
		}
		endLine = startLine + end.span
	}
	if startLine > 0 && endLine > 0 && endLine <= startLine {
		return fmt.Errorf("grid-column end line must be greater than start line")
	}
	panel.Base().Layout().SetGridColumn(startLine, endLine)
	return nil
}
