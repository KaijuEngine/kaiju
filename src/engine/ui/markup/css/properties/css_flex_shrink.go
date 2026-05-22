/******************************************************************************/
/* css_flex_shrink.go                                                         */
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

func (p FlexShrink) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	if values[0].Str == "initial" || values[0].Str == "inherit" || values[0].Str == "unset" {
		panel.Base().Layout().SetFlexShrink(1)
		return nil
	}
	shrink, ok := parseFlexFloat(values[0].Str)
	if !ok {
		return fmt.Errorf("invalid flex-shrink value %q", values[0].Str)
	}
	panel.Base().Layout().SetFlexShrink(shrink)
	return nil
}
