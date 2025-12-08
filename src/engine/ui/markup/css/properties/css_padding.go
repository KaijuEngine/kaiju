/******************************************************************************/
/* css_padding.go                                                             */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"errors"
	"slices"

	"github.com/KaijuEngine/kaiju/engine"
	"github.com/KaijuEngine/kaiju/engine/ui"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/css/helpers"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/css/rules"
	"github.com/KaijuEngine/kaiju/engine/ui/markup/document"
	"github.com/KaijuEngine/kaiju/matrix"
	"github.com/KaijuEngine/kaiju/platform/windowing"
)

func paddingSizeFromString(elm *document.Element, str string, idx matrix.VectorComponent, window *windowing.Window) (matrix.Vec4, error) {
	current := elm.UI.Layout().Padding()
	size := current[idx]
	switch str {
	case "initial":
		size = 0
	case "inherit":
		if elm.Parent.Value() == nil {
			size = 0
		} else {
			size = elm.Parent.Value().UI.Layout().Padding()[idx]
		}
	default:
		size = helpers.NumFromLength(str, window)
	}
	current[idx] = size
	return current, nil
}

func (Padding) Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	switch len(values) {
	case 1:
		for i := range 3 {
			values = append(values, values[i])
		}
	case 2:
		values = append(values, values[0])
		values = append(values, values[1])
	case 3:
		values = append(values, values[1])
	}
	for i := 1; i < len(rules); i++ {
		removeRule := false
		switch rules[i].Property {
		case "padding-top":
			values[0] = rules[i].Values[0]
			removeRule = true
		case "padding-right":
			values[1] = rules[i].Values[0]
			removeRule = true
		case "padding-bottom":
			values[2] = rules[i].Values[0]
			removeRule = true
		case "padding-left":
			values[3] = rules[i].Values[0]
			removeRule = true
		}
		if removeRule {
			rules = slices.Delete(rules, i, i+1)
			i--
		}
	}
	return values, rules
}

// length|initial|inherit
func (Padding) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	var err error
	if len(values) == 1 {
		// all
		err = PaddingLeft{}.Process(panel, elm, values, host)
		err = PaddingRight{}.Process(panel, elm, values, host)
		err = PaddingTop{}.Process(panel, elm, values, host)
		err = PaddingBottom{}.Process(panel, elm, values, host)
	} else if len(values) == 2 {
		// top/bottom, left/right
		err = PaddingTop{}.Process(panel, elm, values[:1], host)
		err = PaddingBottom{}.Process(panel, elm, values[:1], host)
		err = PaddingLeft{}.Process(panel, elm, values[1:], host)
		err = PaddingRight{}.Process(panel, elm, values[1:], host)
	} else if len(values) == 3 {
		// top, left/right, bottom
		err = PaddingTop{}.Process(panel, elm, values[:1], host)
		err = PaddingLeft{}.Process(panel, elm, values[1:2], host)
		err = PaddingRight{}.Process(panel, elm, values[1:2], host)
		err = PaddingBottom{}.Process(panel, elm, values[2:], host)
	} else if len(values) == 4 {
		// top, right, bottom, left
		err = PaddingTop{}.Process(panel, elm, values[:1], host)
		err = PaddingRight{}.Process(panel, elm, values[1:2], host)
		err = PaddingBottom{}.Process(panel, elm, values[2:3], host)
		err = PaddingLeft{}.Process(panel, elm, values[3:], host)
	} else {
		err = errors.New("Padding: Expecting 1-4 values")
	}
	return err
}
