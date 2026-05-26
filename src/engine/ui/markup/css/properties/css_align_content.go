/******************************************************************************/
/* css_align_content.go                                                       */
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

func (p AlignContent) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	switch values[0].Str {
	case "normal", "stretch", "initial", "inherit", "unset":
		panel.SetFlexAlignContent(ui.FlexAlignContentStretch)
	case "start", "flex-start":
		panel.SetFlexAlignContent(ui.FlexAlignContentStart)
	case "end", "flex-end":
		panel.SetFlexAlignContent(ui.FlexAlignContentEnd)
	case "center":
		panel.SetFlexAlignContent(ui.FlexAlignContentCenter)
	case "space-between":
		panel.SetFlexAlignContent(ui.FlexAlignContentSpaceBetween)
	case "space-around":
		panel.SetFlexAlignContent(ui.FlexAlignContentSpaceAround)
	case "space-evenly":
		panel.SetFlexAlignContent(ui.FlexAlignContentSpaceEvenly)
	default:
		return fmt.Errorf("invalid align-content value %q", values[0].Str)
	}
	return nil
}
