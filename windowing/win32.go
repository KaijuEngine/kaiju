//go:build windows

package windowing

import (
	"unicode/utf16"
	"unsafe"
)

/*
#include "windowing.h"

void window_swap_buffers(void* handle) {
	HWND hwnd = (HWND)handle;
	HDC hdc = GetDC(hwnd);
	SwapBuffers(hdc);
}
*/
import "C"

func (e evtMem) toEventType() eventType {
	switch e.EventType() {
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
	default:
		return evtUnknown
	}
}

func createWindowContext(handle unsafe.Pointer, evtSharedMem *evtMem) {
	C.window_create_gl_context(handle, evtSharedMem.AsPointer(), evtSharedMemSize)
}

func createWindow(windowName string, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	go C.window_main(title, evtSharedMem.AsPointer(), evtSharedMemSize)
}

func swapBuffers(handle unsafe.Pointer) {
	C.window_swap_buffers(handle)
}
