/******************************************************************************/
/* css_margin.go                                                              */
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
	"errors"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/helpers"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/windowing"
	"slices"
)

func marginSizeFromStr(str string, window *windowing.Window) float32 {
	if val, ok := borderSizes[str]; ok {
		return val
	} else {
		return helpers.NumFromLength(str, window)
	}
}

func preprocLeftTopRightBottom(values []rules.PropertyValue, rules []rules.Rule, propName string) ([]rules.PropertyValue, []rules.Rule) {
	for i := 1; i < len(rules); i++ {
		if rules[i].Property == propName {
			return values, rules[1:]
		}
	}
	return values, rules
}

func (Margin) Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
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
		case "margin-top":
			values[0] = rules[i].Values[0]
			removeRule = true
		case "margin-right":
			values[1] = rules[i].Values[0]
			removeRule = true
		case "margin-bottom":
			values[2] = rules[i].Values[0]
			removeRule = true
		case "margin-left":
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

func (Margin) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	var err error
	if len(values) == 1 {
		err = MarginLeft{}.Process(panel, elm, values, host)
		err = MarginTop{}.Process(panel, elm, values, host)
		err = MarginRight{}.Process(panel, elm, values, host)
		err = MarginBottom{}.Process(panel, elm, values, host)
	} else if len(values) == 2 {
		err = MarginTop{}.Process(panel, elm, values[:1], host)
		err = MarginBottom{}.Process(panel, elm, values[:1], host)
		err = MarginLeft{}.Process(panel, elm, values[1:], host)
		err = MarginRight{}.Process(panel, elm, values[1:], host)
	} else if len(values) == 3 {
		err = MarginTop{}.Process(panel, elm, values[:1], host)
		err = MarginRight{}.Process(panel, elm, values[1:2], host)
		err = MarginLeft{}.Process(panel, elm, values[1:2], host)
		err = MarginBottom{}.Process(panel, elm, values[2:], host)
	} else if len(values) == 4 {
		err = MarginTop{}.Process(panel, elm, values[:1], host)
		err = MarginRight{}.Process(panel, elm, values[1:2], host)
		err = MarginBottom{}.Process(panel, elm, values[2:3], host)
		err = MarginLeft{}.Process(panel, elm, values[3:], host)
	} else {
		err = errors.New("Margin requires 1-4 values")
	}
	return err
}
