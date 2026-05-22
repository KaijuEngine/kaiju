/******************************************************************************/
/* css_box_sizing.go                                                          */
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

func (p BoxSizing) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("BoxSizing requires exactly 1 value")
	}
	switch values[0].Str {
	case "border-box":
		enableBorderBoxSizing(panel)
	case "content-box", "initial":
		enableContentBoxSizing(panel)
	case "inherit":
		if elm.Parent.Value() != nil && elm.Parent.Value().UI != nil {
			parentPanel := elm.Parent.Value().UI.ToPanel()
			parent := currentSizingConstraints(parentPanel)
			if parent.UsesBorderBox() {
				enableBorderBoxSizing(panel)
			} else {
				enableContentBoxSizing(panel)
			}
		}
	default:
		return fmt.Errorf("unsupported box-sizing value: %s", values[0].Str)
	}
	return nil
}
