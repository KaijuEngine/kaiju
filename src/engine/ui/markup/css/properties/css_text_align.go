/******************************************************************************/
/* css_text_align.go                                                          */
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
	"kaijuengine.com/rendering"
)

// left|right|center|justify|initial|inherit
func (p TextAlign) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	labels := childLabels(elm)
	switch values[0].Str {
	case "left":
		for _, l := range labels {
			l.SetJustify(rendering.FontJustifyLeft)
		}
	case "right":
		for _, l := range labels {
			l.SetJustify(rendering.FontJustifyRight)
		}
	case "center":
		for _, l := range labels {
			l.SetJustify(rendering.FontJustifyCenter)
		}
	case "justify":
		for _, l := range labels {
			l.SetJustify(rendering.FontJustifyJustify)
		}
	case "initial":
		for _, l := range labels {
			l.SetJustify(rendering.FontJustifyLeft)
		}
	case "inherit":
		inherited := rendering.FontJustifyLeft
		if parent := elm.Parent.Value(); parent != nil {
			if parentLabels := childLabels(parent); len(parentLabels) > 0 {
				inherited = parentLabels[0].Justify()
			}
		}
		for _, l := range labels {
			l.SetJustify(inherited)
		}
	}

	return nil
}
