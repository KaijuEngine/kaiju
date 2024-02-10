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

func asEventType(msg int, e *evtMem) eventType {
	switch msg {
	case 2:
		return evtKeyDown
	case 3:
		return evtKeyUp
	case 6:
		return evtMouseMove
	case 4:
		switch e.toMouseEvent().buttonId {
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
		switch e.toMouseEvent().buttonId {
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
	C.window_main(title, C.int(width), C.int(height), evtSharedMem.AsPointer(), evtSharedMemSize)
}

func (w *Window) showWindow(evtSharedMem *evtMem) {
	C.window_show(w.handle)
}

func (w *Window) destroy() {
	C.window_destroy(w.handle)
}

func (w *Window) poll() {
	//evtType := uint32(C.window_poll_controller(w.handle))
	//if evtType != 0 {
	//	w.processControllerEvent(asEventType(evtType))
	//}
	evtType := 1
	for evtType != 0 && !w.evtSharedMem.IsQuit() {
		evtType = int(C.window_poll(w.handle))
		if evtType != 0 {
			t := asEventType(evtType, w.evtSharedMem)
			w.processEvent(t)
		}
	}
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

func (w *Window) getDPI() (int, int, error) {
	klib.NotYetImplemented(131)
	return 96, 96, nil
}

func (w *Window) cHandle() unsafe.Pointer   { return C.window(w.handle) }
func (w *Window) cInstance() unsafe.Pointer { return C.display(w.handle) }
