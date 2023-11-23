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

func toEventType(nativeType uint32) eventType {
	switch nativeType {
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
	default:
		return evtUnknown
	}
}

func createWindow(windowName string, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	go C.window_main(title, evtSharedMem.AsPointer(), evtSharedMemSize)
	for !evtSharedMem.IsReady() {
	}
}
