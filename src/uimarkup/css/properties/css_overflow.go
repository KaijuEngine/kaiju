package properties

import (
	"errors"
	"fmt"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// overflow: visible|hidden|clip|scroll|auto|initial|inherit;
func (p Overflow) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Overflow expects 1 value")
	} else {
		s := values[0].Str
		switch s {
		case "hidden":
			fallthrough
		case "clip":
			panel.GenerateScissor()
			panel.SetScrollDirection(ui.PanelScrollDirectionNone)
		case "scroll":
			fallthrough
		case "auto":
			panel.GenerateScissor()
			panel.SetScrollDirection(ui.PanelScrollDirectionBoth)
		case "inherit":
			if elm.HTML.Parent != nil {
				parentPanel := elm.HTML.Parent.DocumentElement.UI.(*ui.Panel)
				panel.GenerateScissor()
				panel.SetScrollDirection(parentPanel.ScrollDirection())
			}
		case "initial":
			fallthrough
		case "visible":
			panel.DisconnectParentScissor()
		default:
			return fmt.Errorf("Overflow expected a valid value, but got: %s", s)
		}
	}
	return nil
}
