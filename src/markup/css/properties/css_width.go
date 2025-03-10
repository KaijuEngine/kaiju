/******************************************************************************/
/* css_width.go                                                               */
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

package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/markup/css/functions"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
	"strings"
)

func (p Width) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	var width float32
	var err error = nil
	if values[0].Str == "initial" {
		return nil
	}
	if len(values) != 1 {
		err = fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	} else if values[0].Str == "fit-content" {
		panel.FitContentWidth()
		return nil
	} else {
		width = helpers.NumFromLength(values[0].Str, host.Window)
	}
	if err == nil {
		if strings.HasSuffix(values[0].Str, "%") && elm.Parent != nil {
			panel.Base().Layout().AddFunction(func(l *ui.Layout) {
				if l.Ui().Entity().IsRoot() {
					return
				}
				pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
				s := pLayout.PixelSize().X()
				pPad := pLayout.Padding()
				s -= pPad.X() + pPad.Z()
				// Subtracting local padding because it's added in final scale
				p := l.Padding()
				w := s*width - p.X() - p.Z()
				l.ScaleWidth(w)
			})
		} else if values[0].IsFunction() {
			if values[0].Str == "calc" {
				panel.Base().Layout().AddFunction(func(l *ui.Layout) {
					val := values[0]
					val.Args = append(val.Args, "width")
					res, _ := functions.Calc{}.Process(panel, elm, val)
					width = helpers.NumFromLength(res, host.Window)
					l.ScaleWidth(width)
				})
			}
		} else {
			panel.Base().Layout().ScaleWidth(width)
		}
		panel.DontFitContentWidth()
	}
	return err
}
