/******************************************************************************/
/* css_cursor.go                                                              */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).    */
/* Copyright (c) 2015-2023 Brent Farris.                                      */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/klib"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (p Cursor) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return errors.New("no cursor value")
	}
	panel.AddEvent(ui.EventTypeEnter, func() {
		switch values[0].Str {
		case "text":
			host.Window.CursorIbeam()
		case "default":
			host.Window.CursorStandard()
		case "context-menu":
			fallthrough
		case "help":
			fallthrough
		case "pointer":
			fallthrough
		case "progress":
			fallthrough
		case "wait":
			fallthrough
		case "cell":
			fallthrough
		case "crosshair":
			fallthrough
		case "vertical-text":
			fallthrough
		case "alias":
			fallthrough
		case "copy":
			fallthrough
		case "move":
			fallthrough
		case "no-drop":
			fallthrough
		case "not-allowed":
			fallthrough
		case "grab":
			fallthrough
		case "grabbing":
			fallthrough
		case "all-scroll":
			fallthrough
		case "col-resize":
			fallthrough
		case "row-resize":
			fallthrough
		case "n-resize":
			fallthrough
		case "e-resize":
			fallthrough
		case "s-resize":
			fallthrough
		case "w-resize":
			fallthrough
		case "ne-resize":
			fallthrough
		case "nw-resize":
			fallthrough
		case "se-resize":
			fallthrough
		case "sw-resize":
			fallthrough
		case "ew-resize":
			fallthrough
		case "ns-resize":
			fallthrough
		case "nesw-resize":
			fallthrough
		case "nwse-resize":
			fallthrough
		case "zoom-in":
			fallthrough
		case "zoom-out":
			klib.NotYetImplemented(180)
		default:
			host.Window.CursorStandard()
		}
	})
	panel.AddEvent(ui.EventTypeExit, func() {
		host.Window.CursorStandard()
	})
	return nil
}
