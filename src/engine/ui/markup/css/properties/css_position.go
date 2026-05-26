/******************************************************************************/
/* css_position.go                                                            */
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

func (p Position) Sort() int { return -1 }

// static|absolute|fixed|relative|sticky|initial|inherit
func (p Position) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Position requires 1 value")
	}
	switch values[0].Str {
	case "static":
		panel.Base().Layout().SetPositioning(ui.PositioningStatic)
	case "absolute":
		panel.Base().Layout().SetPositioning(ui.PositioningAbsolute)
	case "fixed":
		panel.Base().Layout().SetPositioning(ui.PositioningFixed)
	case "relative":
		panel.Base().Layout().SetPositioning(ui.PositioningRelative)
	case "sticky":
		panel.Base().Layout().SetPositioning(ui.PositioningSticky)
	case "initial":
		panel.Base().Layout().SetPositioning(ui.PositioningStatic)
	case "inherit":
		if elm.Parent.Value() != nil {
			panel.Base().Layout().SetPositioning(elm.Parent.Value().UI.Layout().Positioning())
		}
	default:
		return errors.New("Position invalid position value")
	}
	return nil
}
