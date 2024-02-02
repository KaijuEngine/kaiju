//go:build windows

package windowing

import (
	"kaiju/klib"
	"unicode/utf16"
	"unsafe"
)

/*
#include "windowing.h"
*/
import "C"

func (e evtMem) toEventType() eventType {
	switch e.EventType() {
	case 0x0005:
		return evtResize
	case 0x0104:
		fallthrough
	case 0x0100:
		return evtKeyDown
	case 0x0105:
		fallthrough
	case 0x0101:
		return evtKeyUp
	case 512:
		return evtMouseMove
	case 513:
		return evtLeftMouseDown
	case 514:
		return evtLeftMouseUp
	case 516:
		return evtRightMouseDown
	case 517:
		return evtRightMouseUp
	case 519:
		return evtMiddleMouseDown
	case 520:
		return evtMiddleMouseUp
	case 523:
		return evtX1MouseDown
	case 524:
		return evtX1MouseUp
	case 0x020A:
		return evtMouseWheelVertical
	case 0x020E:
		return evtMouseWheelHorizontal
	case 0xFFFFFFFF - 1:
		return evtControllerStates
	default:
		return evtUnknown
	}
}

func createWindow(windowName string, width, height int, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	go C.window_main(title, C.int(width), C.int(height), evtSharedMem.AsPointer(), evtSharedMemSize)
}

func (w *Window) cursorStandard() {
	C.window_cursor_standard(w.handle)
}

func (w *Window) cursorIbeam() {
	C.window_cursor_ibeam(w.handle)
}

func (w *Window) copyToClipboard(text string) {
	klib.NotYetImplemented(102)
}

func (w *Window) clipboardContents() string {
	klib.NotYetImplemented(102)
	return ""
}
