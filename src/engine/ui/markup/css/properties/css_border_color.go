/******************************************************************************/
/* css_border_color.go                                                        */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/helpers"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
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
