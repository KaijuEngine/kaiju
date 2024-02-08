package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
)

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

func (p BorderColor) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
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
