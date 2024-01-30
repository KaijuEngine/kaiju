//go:build windows

package windowing

import (
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
	case 256:
		return evtKeyDown
	case 0x0105:
		fallthrough
	case 257:
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
