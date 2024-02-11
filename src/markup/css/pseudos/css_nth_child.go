/*****************************************************************************/
/* css_nth_child.go                                                          */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package pseudos

import (
	"errors"
	"fmt"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

func nth(args []string, count int) (int, int, error) {
	if len(args) == 0 {
		return 0, 0, errors.New("no arguments supplied")
	} else if count == 0 {
		return 0, 0, errors.New("no children")
	} else {
		start := 0
		skip := 0
		var err error
		if args[0] == "even" {
			args[0] = "2"
		} else if args[0] == "odd" {
			start = 1
			args[0] = "2"
		}
		helpers.ChangeNToChildCount(args, count)
		if skip, err = helpers.ArithmeticString(args); err != nil {
			return 0, 0, err
		} else if skip <= 0 {
			return 0, 0, fmt.Errorf("invalid skip value: %d", skip)
		}
		return start, skip, nil
	}
}

func (p NthChild) Process(elm document.DocElement, value rules.SelectorPart) ([]document.DocElement, error) {
	if start, skip, err := nth(value.Args, len(elm.HTML.Children)); err == nil {
		selected := make([]document.DocElement, 0)
		for i := start; i < len(elm.HTML.Children); i += skip {
			selected = append(selected, *elm.HTML.Children[i].DocumentElement)
		}
		return selected, nil
	} else {
		return []document.DocElement{}, err
	}
}
