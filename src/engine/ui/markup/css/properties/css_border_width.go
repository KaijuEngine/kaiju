/******************************************************************************/
/* css_border_width.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"slices"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p BorderWidth) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	values = expandFourSideValues(values)
	for i := 1; i < len(ruleList); i++ {
		removeRule := false
		switch ruleList[i].Property {
		case "border-top-width":
			values[0] = ruleList[i].Values[0]
			removeRule = true
		case "border-right-width":
			values[1] = ruleList[i].Values[0]
			removeRule = true
		case "border-bottom-width":
			values[2] = ruleList[i].Values[0]
			removeRule = true
		case "border-left-width":
			values[3] = ruleList[i].Values[0]
			removeRule = true
		}
		if removeRule {
			ruleList = slices.Delete(ruleList, i, i+1)
			i--
		}
	}
	ruleList[0].Values = values
	return values, ruleList
}

func (p BorderWidth) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	b := [4]float32{}
	if len(values) == 1 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = b[0]
		b[2] = b[0]
		b[3] = b[0]
	} else if len(values) == 2 {
		// Top/bottom left/right
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[0] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = b[1]
	} else if len(values) == 3 {
		// Top left/right bottom
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[0] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = helpers.NumFromLength(values[2].Str, host.Window)
	} else if len(values) == 4 {
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[2] = helpers.NumFromLength(values[1].Str, host.Window)
		b[3] = helpers.NumFromLength(values[2].Str, host.Window)
		b[0] = helpers.NumFromLength(values[3].Str, host.Window)
	} else {
		return errors.New("Invalid number of values for BorderRadius")
	}
	panel.SetBorderSize(b[0], b[1], b[2], b[3])
	return nil
}
