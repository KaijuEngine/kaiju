package globals

import "kaiju/systems/events"

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
}
