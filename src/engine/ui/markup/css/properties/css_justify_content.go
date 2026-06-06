/******************************************************************************/
/* css_justify_content.go                                                     */
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

func (p JustifyContent) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	switch values[0].Str {
	case "normal", "start", "flex-start", "left", "initial", "inherit", "unset":
		panel.SetFlexJustify(ui.FlexJustifyStart)
	case "end", "flex-end", "right":
		panel.SetFlexJustify(ui.FlexJustifyEnd)
	case "center":
		panel.SetFlexJustify(ui.FlexJustifyCenter)
	case "space-between":
		panel.SetFlexJustify(ui.FlexJustifySpaceBetween)
	case "space-around":
		panel.SetFlexJustify(ui.FlexJustifySpaceAround)
	case "space-evenly":
		panel.SetFlexJustify(ui.FlexJustifySpaceEvenly)
	default:
		return fmt.Errorf("invalid justify-content value %q", values[0].Str)
	}
	return nil
}
