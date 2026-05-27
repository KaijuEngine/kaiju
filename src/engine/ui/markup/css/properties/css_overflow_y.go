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
		direction := panel.ScrollDirection() &^ ui.PanelScrollDirectionVertical
		if direction == ui.PanelScrollDirectionNone {
			panel.SetOverflow(ui.OverflowHidden)
		} else {
			panel.SetOverflow(ui.OverflowScroll)
		}
		panel.SetScrollDirection(direction)
		panel.Base().GenerateScissor()
	case "auto":
		fallthrough
	case "scroll":
		panel.SetScrollDirection(panel.ScrollDirection() | ui.PanelScrollDirectionVertical)
		panel.SetOverflow(ui.OverflowScroll)
		panel.Base().GenerateScissor()
		panel.DontFitContentHeight()
		panel.DontFitContentHeight()
	case "inherit":
		if elm.Parent.Value() != nil {
			parentPanel := elm.Parent.Value().UIPanel
			panel.SetOverflow(parentPanel.Overflow())
			panel.SetScrollDirection(parentPanel.ScrollDirection() | ui.PanelScrollDirectionVertical)
			panel.Base().GenerateScissor()
		}
	case "initial":
		fallthrough
	case "visible":
		direction := panel.ScrollDirection() &^ ui.PanelScrollDirectionVertical
		if direction == ui.PanelScrollDirectionNone {
			panel.SetOverflow(ui.OverflowVisible)
		} else {
			panel.SetOverflow(ui.OverflowScroll)
		}
		panel.SetScrollDirection(direction)
	default:
		return fmt.Errorf("OverflowX expected a valid value, but got: %s", s)
	}

	return nil
}
