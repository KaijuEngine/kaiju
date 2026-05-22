/******************************************************************************/
/* css_display.go                                                             */
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

// block|inline|inline-block|flex|inline-flex|grid|inline-grid|flow-root|none|contents|block flex|block flow|block flow-root|block grid|inline flex|inline flow|inline flow-root|inline grid|table|table-row|list-item|inherit|initial|revert|revert-layer|unset
func (p Display) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return errors.New("no values for display")
	}
	switch values[0].Str {
	case "none":
		panel.Base().Hide()
		return nil
	case "flex", "inline-flex", "block flex", "inline flex":
		panel.SetFlex()
		return nil
	case "grid", "inline-grid", "block grid", "inline grid":
		panel.SetGrid(0)
		return nil
	case "block", "inline", "inline-block", "flow-root", "block flow", "inline flow":
		panel.SetFlowLayout()
		return nil
	default:
		return nil
	}
}
