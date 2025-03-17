package windowing

import (
	"kaiju/platform/hid"
	"kaiju/engine/systems/events"
)

var (
	dragData     DragDataTarget
	OnDragStop   events.Event
	dragDataUsed bool
)

type DragDataTarget interface {
	DragUpdate()
}

func HasDragData() bool        { return dragData != nil }
func DragData() DragDataTarget { return dragData }
func IsDragDataUsed() bool     { return dragDataUsed }

func UseDragData() DragDataTarget {
	dragDataUsed = true
	return dragData
}

func SetDragData(data DragDataTarget) {
	if dragData != nil {
		OnDragStop.Execute()
	}
	dragData = data
	dragDataUsed = false
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
