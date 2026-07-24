/******************************************************************************/
/* css_flex.go                                                                */
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
	"kaijuengine.com/matrix"
)

func (p Flex) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	layout := panel.Base().Layout()
	if len(values) == 1 {
		switch values[0].Str {
		case "none":
			layout.SetFlexGrow(0)
			layout.SetFlexShrink(0)
			layout.SetFlexBasisAuto()
			return nil
		case "auto":
			layout.SetFlexGrow(1)
			layout.SetFlexShrink(1)
			layout.SetFlexBasisAuto()
			return nil
		case "initial", "inherit", "unset":
			layout.SetFlexGrow(0)
			layout.SetFlexShrink(1)
			layout.SetFlexBasisAuto()
			return nil
		}
		if grow, ok := parseFlexFloat(values[0].Str); ok {
			layout.SetFlexGrow(grow)
			layout.SetFlexShrink(1)
			layout.SetFlexBasis(0, false)
			if grow > 0 {
				panel.DontFitContent()
			}
			return nil
		}
		layout.SetFlexGrow(0)
		layout.SetFlexShrink(1)
		setFlexBasis(panel, values[0].Str, host)
		return nil
	}
	var grow matrix.Float
	if g, ok := parseFlexFloat(values[0].Str); ok {
		grow = g
		layout.SetFlexGrow(g)
	} else {
		return fmt.Errorf("invalid flex-grow value %q", values[0].Str)
	}
	basisIdx := 1
	if shrink, ok := parseFlexFloat(values[1].Str); ok {
		layout.SetFlexShrink(shrink)
		basisIdx = 2
	} else {
		layout.SetFlexShrink(1)
	}
	if basisIdx < len(values) {
		setFlexBasis(panel, values[basisIdx].Str, host)
	} else {
		layout.SetFlexBasis(0, false)
	}
	if grow > 0 {
		panel.DontFitContent()
	}
	return nil
}
