/******************************************************************************/
/* drag_data.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package windowing

import (
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
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
			w.Mouse.SetPosition(matrix.Float(lx), matrix.Float(ly),
				matrix.Float(w.width), matrix.Float(w.height))
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
			w.Mouse.SetPosition(matrix.Float(lx), matrix.Float(ly),
				matrix.Float(w.width), matrix.Float(w.height))
			w.Mouse.SetUp(hid.MouseButtonLeft)
			w.windowSync <- struct{}{}
		}
	}
}
