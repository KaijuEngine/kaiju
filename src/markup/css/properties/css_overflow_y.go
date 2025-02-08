/******************************************************************************/
/* css_overflow_y.go                                                          */
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
	"fmt"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (p OverflowY) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Overflow expects 1 value")
	} else {
		s := values[0].Str
		switch s {
		case "clip":
			fallthrough
		case "hidden":
			panel.SetOverflow(ui.OverflowHidden)
			panel.Base().GenerateScissor()
			panel.SetScrollDirection(panel.ScrollDirection() ^ ui.PanelScrollDirectionVertical)
		case "auto":
			fallthrough
		case "scroll":
			panel.SetOverflow(ui.OverflowScroll)
			panel.Base().GenerateScissor()
			panel.SetScrollDirection(panel.ScrollDirection() | ui.PanelScrollDirectionVertical)
		case "inherit":
			if elm.Parent != nil {
				parentPanel := elm.Parent.UIPanel
				panel.SetOverflow(parentPanel.Overflow())
				panel.Base().GenerateScissor()
				panel.SetScrollDirection(parentPanel.ScrollDirection() | ui.PanelScrollDirectionVertical)
			}
		case "initial":
			fallthrough
		case "visible":
			panel.SetOverflow(ui.OverflowVisible)
		default:
			return fmt.Errorf("OverflowX expected a valid value, but got: %s", s)
		}
	}
	return nil
}
