/******************************************************************************/
/* css_flex_flow.go                                                           */
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

func (p FlexFlow) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	seen := false
	for i := range values {
		if setFlexDirection(panel, values[i].Str) || setFlexWrap(panel, values[i].Str) {
			seen = true
		} else {
			return fmt.Errorf("invalid flex-flow value %q", values[i].Str)
		}
	}
	if !seen {
		panel.SetFlex()
	}
	return nil
}
