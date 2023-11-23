//go:build linux || darwin

package windowing

/*
#include "windowing.h"
*/
import "C"
import (
	"unsafe"
)

func (e evtMem) toEventType() eventType {
	switch e.EventType() {
	case 6:
		return evtMouseMove
	case 4:
		switch e.toMouseEvent().mouseButtonId {
		case nativeMouseButtonLeft:
			return evtLeftMouseDown
		case nativeMouseButtonMiddle:
			return evtMiddleMouseDown
		case nativeMouseButtonRight:
			return evtRightMouseDown
		case nativeMouseButtonX1:
			return evtX1MouseDown
		case nativeMouseButtonX2:
			return evtX2MouseDown
		default:
			return evtUnknown
		}
	case 5:
		switch e.toMouseEvent().mouseButtonId {
		case nativeMouseButtonLeft:
			return evtLeftMouseUp
		case nativeMouseButtonMiddle:
			return evtMiddleMouseUp
		case nativeMouseButtonRight:
			return evtRightMouseUp
		case nativeMouseButtonX1:
			return evtX1MouseUp
		case nativeMouseButtonX2:
			return evtX2MouseUp
		default:
			return evtUnknown
		}
	default:
		return evtUnknown
	}
}

func createWindow(windowName string, evtSharedMem *evtMem) {
	title := C.CString(string(windowName))
	defer C.free(unsafe.Pointer(title))
	go C.window_main(title, evtSharedMem.AsPointer(), evtSharedMemSize)
	for !evtSharedMem.IsReady() {
	}
}
