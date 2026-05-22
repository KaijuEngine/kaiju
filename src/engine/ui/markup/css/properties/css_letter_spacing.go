/******************************************************************************/
/* css_letter_spacing.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p LetterSpacing) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}
	labels := childLabels(elm)
	spacing := float32(0)
	switch values[0].Str {
	case "normal", "initial", "unset":
	case "inherit":
		if parent := elm.Parent.Value(); parent != nil {
			if parentLabels := childLabels(parent); len(parentLabels) > 0 {
				spacing = parentLabels[0].LetterSpacing()
			}
		}
	default:
		emSize := float32(16)
		if len(labels) > 0 {
			emSize = host.FontCache().EMSize(labels[0].FontFace())
		}
		spacing = helpers.NumFromLengthWithFont(values[0].Str, host.Window, emSize)
	}
	for _, label := range labels {
		label.SetLetterSpacing(spacing)
	}
	return nil
}
