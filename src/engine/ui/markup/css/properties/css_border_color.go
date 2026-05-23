/******************************************************************************/
/* css_border_color.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"
	"slices"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (p BorderColor) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	values = expandFourSideValues(values)
	for i := 1; i < len(ruleList); i++ {
		removeRule := false
		switch ruleList[i].Property {
		case "border-top-color":
			values[0] = ruleList[i].Values[0]
			removeRule = true
		case "border-right-color":
			values[1] = ruleList[i].Values[0]
			removeRule = true
		case "border-bottom-color":
			values[2] = ruleList[i].Values[0]
			removeRule = true
		case "border-left-color":
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

func colorValues(values []rules.PropertyValue) []string {
	hexes := make([]string, len(values))
	for i, v := range values {
		hex := v.Str
		if newHex, ok := helpers.ColorMap[v.Str]; ok {
			hex = newHex
		}
		hexes[i] = hex
	}
	return hexes
}

func (p BorderColor) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	var err error
	colors := [4]matrix.Color{}
	if len(values) == 1 {
		hex := values[0].Str
		if newHex, ok := helpers.ColorMap[hex]; ok {
			hex = newHex
		}
		if colors[0], err = matrix.ColorFromHexString(hex); err == nil {
			colors[1] = colors[0]
			colors[2] = colors[0]
			colors[3] = colors[0]
		}
	} else if len(values) == 2 {
		// Top/bottom left/right
		hexes := colorValues(values)
		if colors[1], err = matrix.ColorFromHexString(hexes[0]); err == nil {
			colors[3] = colors[1]
		}
		if colors[0], err = matrix.ColorFromHexString(hexes[1]); err == nil {
			colors[2] = colors[0]
		}
	} else if len(values) == 3 {
		// Top left/right bottom
		hexes := colorValues(values)
		colors[1], err = matrix.ColorFromHexString(hexes[0])
		if colors[0], err = matrix.ColorFromHexString(hexes[1]); err == nil {
			colors[2] = colors[0]
		}
		colors[2], err = matrix.ColorFromHexString(hexes[2])
	} else if len(values) == 4 {
		// Top right bottom left
		hexes := colorValues(values)
		colors[1], err = matrix.ColorFromHexString(hexes[0])
		colors[2], err = matrix.ColorFromHexString(hexes[1])
		colors[3], err = matrix.ColorFromHexString(hexes[2])
		colors[0], err = matrix.ColorFromHexString(hexes[3])
	} else {
		err = fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	}
	if err == nil {
		panel.SetBorderColor(colors[0], colors[1], colors[2], colors[3])
	}
	return err
}
