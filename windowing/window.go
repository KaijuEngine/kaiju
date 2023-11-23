package windowing

import (
	"kaiju/hid"
)

type eventType int

const (
	evtUnknown eventType = iota
	evtMouseMove
	evtLeftMouseDown
	evtLeftMouseUp
	evtMiddleMouseDown
	evtMiddleMouseUp
	evtRightMouseDown
	evtRightMouseUp
	evtX1MouseDown
	evtX1MouseUp
	evtX2MouseDown
	evtX2MouseUp
	evtMouseWheelVertical
	evtMouseWheelHorizontal
)

type Window struct {
	Mouse         hid.Mouse
	evtSharedMem  evtMem
	width, height int
	isClosed      bool
	isCrashed     bool
}

func New(windowName string) *Window {
	w := &Window{
		Mouse:  hid.NewMouse(),
		width:  1280,
		height: 720,
	}
	// TODO:  Pass in width and height
	createWindow(windowName, &w.evtSharedMem)
	return w
}

func (w Window) IsClosed() bool {
	return w.isClosed
}

func (w Window) IsCrashed() bool {
	return w.isCrashed
}

func (w *Window) processEvent() {
	evtType := w.evtSharedMem.toEventType()
	w.processMouseEvent(evtType)
}

func (w *Window) processMouseEvent(evtType eventType) {
	switch evtType {
	case evtMouseMove:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetPosition(float32(me.mouseX), float32(me.mouseY), float32(w.width), float32(w.height))
	case evtLeftMouseDown:
		w.Mouse.SetDown(hid.MouseButtonLeft)
	case evtLeftMouseUp:
		w.Mouse.SetUp(hid.MouseButtonLeft)
	case evtMiddleMouseDown:
		w.Mouse.SetDown(hid.MouseButtonMiddle)
	case evtMiddleMouseUp:
		w.Mouse.SetUp(hid.MouseButtonMiddle)
	case evtRightMouseDown:
		w.Mouse.SetDown(hid.MouseButtonRight)
	case evtRightMouseUp:
		w.Mouse.SetUp(hid.MouseButtonRight)
	case evtX1MouseDown:
		me := w.evtSharedMem.toMouseEvent()
		if me.mouseButtonId == 4 {
			println("X2 down")
			w.Mouse.SetDown(hid.MouseButtonX2)
		} else {
			println("X1 down")
			w.Mouse.SetDown(hid.MouseButtonX1)
		}
	case evtX1MouseUp:
		me := w.evtSharedMem.toMouseEvent()
		if me.mouseButtonId == 4 {
			println("X2 up")
			w.Mouse.SetUp(hid.MouseButtonX2)
		} else {
			println("X1 up")
			w.Mouse.SetUp(hid.MouseButtonX1)
		}
	case evtX2MouseDown:
		w.Mouse.SetDown(hid.MouseButtonX2)
	case evtX2MouseUp:
		w.Mouse.SetUp(hid.MouseButtonX2)
	case evtMouseWheelVertical:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetScroll(0.0, float32(me.mouseY))
	case evtMouseWheelHorizontal:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetScroll(float32(me.mouseX), 0.0)
	}
}

func (w *Window) Poll() {
	w.evtSharedMem.MakeAvailable()
	for !w.evtSharedMem.IsQuit() && !w.evtSharedMem.IsFatal() {
		for !w.evtSharedMem.IsReady() {
		}
		if w.evtSharedMem.IsWritten() {
			if w.evtSharedMem.HasEvent() {
				w.processEvent()
				w.evtSharedMem.MakeAvailable()
			} else {
				break
			}
		}
	}
	w.isClosed = w.isClosed || w.evtSharedMem.IsQuit()
	w.isCrashed = w.isCrashed || w.evtSharedMem.IsFatal()
}
