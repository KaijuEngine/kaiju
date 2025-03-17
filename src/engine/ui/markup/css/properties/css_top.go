/******************************************************************************/
/* css_top.go                                                                 */
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
	"errors"
	"kaiju/engine"
	"kaiju/engine/ui/markup/css/helpers"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"kaiju/engine/ui"
	"strings"
)

// auto|length|initial|inherit
func (p Top) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("top expects 1 value")
	} else {
		offset := panel.Base().Layout().InnerOffset()
		s := values[0].Str
		layout := elm.UI.Layout()
		switch s {
		case "auto":
		case "initial":
		case "inherit":
			if elm.Parent.Value() != nil {
				offset.SetTop(elm.Parent.Value().UI.Layout().Offset().Y())
			}
		default:
			val := helpers.NumFromLength(values[0].Str, host.Window)
			if strings.HasSuffix(values[0].Str, "%") {
				panel.Base().Layout().AddFunction(func(l *ui.Layout) {
					if l.Ui().Entity().IsRoot() {
						return
					}
					pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
					l.SetInnerOffsetTop(pLayout.PixelSize().Y() * -val)
				})
			} else {
				if layout.Anchor() <= ui.AnchorTopRight {
					val = -val
				}
				offset[matrix.Vy] += val
			}
		}
		layout.SetInnerOffset(offset.X(), offset.Y(), offset.Z(), offset.W())
		layout.AnchorTo(layout.Anchor().ConvertToTop())
	}
	return nil
}
