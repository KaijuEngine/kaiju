/******************************************************************************/
/* css_calc.go                                                                */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

package functions

import (
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
	"strconv"
	"strings"
)

type calcOp int

const (
	calcOpNone = calcOp(iota)
	calcOpAdd
	calcOpSub
	calcOpMul
	calcOpDiv
)

type calcEntry struct {
	value float32
	op    calcOp
}

func (f Calc) Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error) {
	prop := value.Args[len(value.Args)-1]
	value.Args = value.Args[:len(value.Args)-1]
	entries := make([]calcEntry, len(value.Args))
	for i := range value.Args {
		switch value.Args[i] {
		case "+":
			entries[i].op = calcOpAdd
		case "-":
			entries[i].op = calcOpSub
		case "*":
			entries[i].op = calcOpMul
		case "/":
			entries[i].op = calcOpDiv
		default:
			v := helpers.NumFromLength(value.Args[i], panel.Host().Window)
			if strings.HasSuffix(value.Args[i], "%") {
				if prop == "width" {
					pl := elm.Parent.UI.Layout()
					p := pl.Padding()
					v *= pl.PixelSize().Width() - p.X() - p.Z()
				} else if prop == "height" {
					pl := elm.Parent.UI.Layout()
					p := pl.Padding()
					v *= pl.PixelSize().Height() - p.Y() - p.W()
				}
			}
			entries[i].value = v
		}
	}
	// Go through and do all the multiply and divide
	for i := range entries {
		if entries[i].op == calcOpMul {
			entries[i-1].value *= entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		} else if entries[i].op == calcOpDiv {
			entries[i-1].value /= entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		}
	}
	// Go through and do all the add and subtract
	for i := 0; i < len(entries); i++ {
		if entries[i].op == calcOpAdd {
			entries[i-1].value += entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		} else if entries[i].op == calcOpSub {
			entries[i-1].value -= entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		}
	}
	return strconv.FormatFloat(float64(entries[0].value), 'f', 5, 32) + "px", nil
}
