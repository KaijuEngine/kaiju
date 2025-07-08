/******************************************************************************/
/* css_border.go                                                              */
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
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/helpers"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/windowing"
	"strings"
)

var borderSizes = map[string]float32{
	"medium": 2,
	"thin":   1,
	"thick":  4,
}

func borderSizeFromStr(str string, window *windowing.Window, fallback float32) float32 {
	if val, ok := borderSizes[str]; ok {
		return val
	} else if strings.HasSuffix(str, "px") {
		return helpers.NumFromLength(str, window)
	} else {
		return fallback
	}
}

var borderStyleMap = map[string]ui.BorderStyle{
	"none":   ui.BorderStyleNone,
	"hidden": ui.BorderStyleHidden,
	"dotted": ui.BorderStyleDotted,
	"dashed": ui.BorderStyleDashed,
	"solid":  ui.BorderStyleSolid,
	"double": ui.BorderStyleDouble,
	"groove": ui.BorderStyleGroove,
	"ridge":  ui.BorderStyleRidge,
	"inset":  ui.BorderStyleInset,
	"outset": ui.BorderStyleOutset,
}

func borderStyleFromStr(str string, lrtb int, elm *document.Element) (ui.BorderStyle, bool) {
	if val, ok := borderStyleMap[str]; ok {
		return val, true
	} else if str == "initial" {
		// TODO:  Based on tag
		return ui.BorderStyleNone, true
	} else if str == "inherit" && elm.Parent.Value() != nil {
		return elm.Parent.Value().UI.ToPanel().BorderStyle()[lrtb], true
	} else {
		return ui.BorderStyleNone, false
	}
}

// border-width border-style border-color|initial|inherit
func (p Border) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	BorderLeftWidth{}.Process(panel, elm, values[:1], host)
	BorderTopWidth{}.Process(panel, elm, values[:1], host)
	BorderRightWidth{}.Process(panel, elm, values[:1], host)
	BorderBottomWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderLeftStyle{}.Process(panel, elm, values[1:2], host)
		BorderTopStyle{}.Process(panel, elm, values[1:2], host)
		BorderRightStyle{}.Process(panel, elm, values[1:2], host)
		BorderBottomStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderLeftColor{}.Process(panel, elm, values[2:], host)
		BorderTopColor{}.Process(panel, elm, values[2:], host)
		BorderRightColor{}.Process(panel, elm, values[2:], host)
		BorderBottomColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
