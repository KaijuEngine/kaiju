/******************************************************************************/
/* css_flex.go                                                                */
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

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
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
			return nil
		}
		layout.SetFlexGrow(0)
		layout.SetFlexShrink(1)
		setFlexBasis(layout, values[0].Str, host)
		return nil
	}
	if grow, ok := parseFlexFloat(values[0].Str); ok {
		layout.SetFlexGrow(grow)
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
		setFlexBasis(layout, values[basisIdx].Str, host)
	} else {
		layout.SetFlexBasis(0, false)
	}
	return nil
}
