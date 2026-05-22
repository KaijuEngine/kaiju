/******************************************************************************/
/* css_cursor.go                                                              */
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

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/windowing"
)

func (p Cursor) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return errors.New("no cursor value")
	}

	panel.Base().AddEvent(ui.EventTypeEnter, func() {
		switch values[0].Str {
		case "auto":
			host.Window.SetCursor(windowing.CursorKindAuto)
		case "default":
			host.Window.SetCursor(windowing.CursorKindDefault)
		case "none":
			host.Window.SetCursor(windowing.CursorKindNone)
		case "context-menu":
			host.Window.SetCursor(windowing.CursorKindContextMenu)
		case "text":
			host.Window.SetCursor(windowing.CursorKindText)
		case "vertical-text":
			host.Window.SetCursor(windowing.CursorKindVerticalText)
		case "pointer":
			host.Window.SetCursor(windowing.CursorKindPointer)
		case "help":
			host.Window.SetCursor(windowing.CursorKindHelp)
		case "progress":
			host.Window.SetCursor(windowing.CursorKindProgress)
		case "wait":
			host.Window.SetCursor(windowing.CursorKindWait)
		case "cell":
			host.Window.SetCursor(windowing.CursorKindCell)
		case "crosshair":
			host.Window.SetCursor(windowing.CursorKindCrosshair)
		case "alias":
			host.Window.SetCursor(windowing.CursorKindAlias)
		case "copy":
			host.Window.SetCursor(windowing.CursorKindCopy)
		case "move":
			host.Window.SetCursor(windowing.CursorKindMove)
		case "no-drop":
			host.Window.SetCursor(windowing.CursorKindNoDrop)
		case "not-allowed":
			host.Window.SetCursor(windowing.CursorKindNotAllowed)
		case "grab":
			host.Window.SetCursor(windowing.CursorKindGrab)
		case "grabbing":
			host.Window.SetCursor(windowing.CursorKindGrabbing)
		case "all-scroll":
			host.Window.SetCursor(windowing.CursorKindResizeAll)
		case "col-resize":
			host.Window.SetCursor(windowing.CursorKindResizeCol)
		case "e-resize":
			host.Window.SetCursor(windowing.CursorKindResizeE)
		case "w-resize":
			host.Window.SetCursor(windowing.CursorKindResizeW)
		case "ew-resize":
			host.Window.SetCursor(windowing.CursorKindResizeEW)
		case "row-resize":
			host.Window.SetCursor(windowing.CursorKindResizeRow)
		case "n-resize":
			host.Window.SetCursor(windowing.CursorKindResizeN)
		case "s-resize":
			host.Window.SetCursor(windowing.CursorKindResizeS)
		case "ns-resize":
			host.Window.SetCursor(windowing.CursorKindResizeNS)
		case "ne-resize":
			host.Window.SetCursor(windowing.CursorKindResizeNE)
		case "sw-resize":
			host.Window.SetCursor(windowing.CursorKindResizeSW)
		case "nesw-resize":
			host.Window.SetCursor(windowing.CursorKindResizeNESW)
		case "nw-resize":
			host.Window.SetCursor(windowing.CursorKindResizeNW)
		case "se-resize":
			host.Window.SetCursor(windowing.CursorKindResizeSE)
		case "nwse-resize":
			host.Window.SetCursor(windowing.CursorKindResizeNWSE)
		case "zoom-in":
			host.Window.SetCursor(windowing.CursorKindZoomIn)
		case "zoom-out":
			host.Window.SetCursor(windowing.CursorKindZoomOut)
		default:
			host.Window.SetCursor(windowing.CursorKindDefault)
		}
	})

	panel.Base().AddEvent(ui.EventTypeExit, func() {
		host.Window.SetCursor(windowing.CursorKindDefault)
	})

	return nil
}
