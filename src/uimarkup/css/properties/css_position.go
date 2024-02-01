package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// static|absolute|fixed|relative|sticky|initial|inherit
func (p Position) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Position requires 1 value")
	} else {
		var err error
		switch values[0].Str {
		case "static":
			panel.Layout().SetPositioning(ui.PositioningStatic)
		case "absolute":
			panel.Layout().SetPositioning(ui.PositioningAbsolute)
		case "fixed":
			panel.Layout().SetPositioning(ui.PositioningFixed)
		case "relative":
			panel.Layout().SetPositioning(ui.PositioningRelative)
		case "sticky":
			panel.Layout().SetPositioning(ui.PositioningSticky)
		case "initial":
			panel.Layout().SetPositioning(ui.PositioningStatic)
		case "inherit":
			if elm.HTML.Parent != nil {
				panel.Layout().SetPositioning(elm.HTML.DocumentElement.UI.Layout().Positioning())
			}
		default:
			err = errors.New("Position invalid position value")
		}
		return err
	}
}
