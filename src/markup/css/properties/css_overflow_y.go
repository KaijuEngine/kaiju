package properties

import (
	"errors"
	"fmt"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (p OverflowY) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Overflow expects 1 value")
	} else {
		s := values[0].Str
		switch s {
		case "clip":
			fallthrough
		case "hidden":
			panel.SetOverflow(ui.OverflowHidden)
			panel.GenerateScissor()
			panel.SetScrollDirection(panel.ScrollDirection() ^ ui.PanelScrollDirectionVertical)
		case "auto":
			fallthrough
		case "scroll":
			panel.SetOverflow(ui.OverflowScroll)
			panel.GenerateScissor()
			panel.SetScrollDirection(panel.ScrollDirection() | ui.PanelScrollDirectionVertical)
		case "inherit":
			if elm.HTML.Parent != nil {
				parentPanel := elm.HTML.Parent.DocumentElement.UIPanel
				panel.SetOverflow(parentPanel.Overflow())
				panel.GenerateScissor()
				panel.SetScrollDirection(parentPanel.ScrollDirection() | ui.PanelScrollDirectionVertical)
			}
		case "initial":
			fallthrough
		case "visible":
			panel.SetOverflow(ui.OverflowVisible)
		default:
			return fmt.Errorf("OverflowX expected a valid value, but got: %s", s)
		}
	}
	return nil
}
