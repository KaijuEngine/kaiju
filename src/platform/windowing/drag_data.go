/******************************************************************************/
/* drag_data.go                                                               */
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

package windowing

import (
	"kaiju/engine/systems/events"
	"kaiju/platform/hid"
)

var (
	dragData   DragDataTarget
	OnDragStop events.Event
)

type DragDataTarget interface {
	DragUpdate()
}

func HasDragData() bool        { return dragData != nil }
func DragData() DragDataTarget { return dragData }

func SetDragData(data DragDataTarget) {
	if dragData != nil {
		OnDragStop.Execute()
	}
	dragData = data
	OnDragStop.Clear()
}

func UpdateDragData(sender *Window, x, y int) {
	if dragData == nil {
		return
	}
	sx, sy := sender.ToScreenPosition(x, y)
	dragData.DragUpdate()
	if w, ok := FindWindowAtPoint(sx, sy); ok {
		if w != sender {
			lx, ly := w.ToLocalPosition(sx, sy)
			if !w.Mouse.Held(hid.MouseButtonLeft) {
				w.Mouse.ForceHeld(hid.MouseButtonLeft)
			}
			w.Mouse.SetPosition(float32(lx), float32(ly),
				float32(w.width), float32(w.height))
		}
	}
}

func UpdateDragDrop(sender *Window, x, y int) {
	if dragData == nil {
		return
	}
	sx, sy := sender.ToScreenPosition(x, y)
	if w, ok := FindWindowAtPoint(sx, sy); ok {
		if w != sender {
			w.requestSync()
			<-w.windowSync
			lx, ly := w.ToLocalPosition(sx, sy)
			w.Mouse.SetPosition(float32(lx), float32(ly),
				float32(w.width), float32(w.height))
			w.Mouse.SetUp(hid.MouseButtonLeft)
			w.windowSync <- struct{}{}
		}
	}
}
