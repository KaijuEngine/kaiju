//go:build linux || darwin

package windowing

/*
#include <stdlib.h>
#include "windowing.h"
*/
import "C"
import (
	"kaiju/klib"
	"unsafe"
)

func (e evtMem) toEventType() eventType {
	switch e.EventType() {
	case 2:
		return evtKeyDown
	case 3:
		return evtKeyUp
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

func scaleScrollDelta(delta float32) float32 {
	return delta
}

func createWindow(windowName string, width, height int, evtSharedMem *evtMem) {
	title := C.CString(string(windowName))
	defer C.free(unsafe.Pointer(title))
	go C.window_main(title, C.int(width), C.int(height), evtSharedMem.AsPointer(), evtSharedMemSize)
	evtSharedMem.AwaitReady()
}

func (w *Window) cursorStandard() {
	klib.NotYetImplemented(100)
}

func (w *Window) cursorIbeam() {
	klib.NotYetImplemented(101)
}

func (w *Window) copyToClipboard(text string) {
	klib.NotYetImplemented(103)
}

func (w *Window) clipboardContents() string {
	klib.NotYetImplemented(103)
	return ""
}
