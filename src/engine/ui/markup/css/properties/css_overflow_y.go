/******************************************************************************/
/* css_overflow_y.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p OverflowY) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Overflow expects 1 value")
	}

	s := values[0].Str
	switch s {
	case "clip":
		fallthrough
	case "hidden":
		panel.SetOverflow(ui.OverflowHidden)
		panel.Base().GenerateScissor()
		panel.SetScrollDirection(panel.ScrollDirection() ^ ui.PanelScrollDirectionVertical)
	case "auto":
		fallthrough
	case "scroll":
		panel.SetOverflow(ui.OverflowScroll)
		panel.Base().GenerateScissor()
		panel.SetScrollDirection(panel.ScrollDirection() | ui.PanelScrollDirectionVertical)
	case "inherit":
		if elm.Parent.Value() != nil {
			parentPanel := elm.Parent.Value().UIPanel
			panel.SetOverflow(parentPanel.Overflow())
			panel.Base().GenerateScissor()
			panel.SetScrollDirection(parentPanel.ScrollDirection() | ui.PanelScrollDirectionVertical)
		}
	case "initial":
		fallthrough
	case "visible":
		panel.SetOverflow(ui.OverflowVisible)
	default:
		return fmt.Errorf("OverflowX expected a valid value, but got: %s", s)
	}

	return nil
}
