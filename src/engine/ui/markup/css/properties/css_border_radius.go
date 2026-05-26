/******************************************************************************/
/* css_border_radius.go                                                       */
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

func (p BorderRadius) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	values = expandFourSideValues(values)
	for i := 1; i < len(ruleList); i++ {
		removeRule := false
		switch ruleList[i].Property {
		case "border-top-left-radius":
			values[0] = ruleList[i].Values[0]
			removeRule = true
		case "border-top-right-radius":
			values[1] = ruleList[i].Values[0]
			removeRule = true
		case "border-bottom-right-radius":
			values[2] = ruleList[i].Values[0]
			removeRule = true
		case "border-bottom-left-radius":
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

func (p BorderRadius) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	b := [4]float32{}
	if len(values) == 1 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = b[0]
		b[2] = b[0]
		b[3] = b[0]
	} else if len(values) == 2 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = b[1]
	} else if len(values) == 3 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = helpers.NumFromLength(values[2].Str, host.Window)
		b[3] = b[1]
	} else if len(values) == 4 {
		for i := 0; i < 4; i++ {
			b[i] = helpers.NumFromLength(values[i].Str, host.Window)
		}
	} else {
		return errors.New("Invalid number of values for BorderRadius")
	}
	panel.SetBorderRadius(b[0], b[1], b[2], b[3])
	return nil
}
