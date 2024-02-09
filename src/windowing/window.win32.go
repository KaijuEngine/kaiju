//go:build windows

package windowing

import (
	"kaiju/klib"
	"strings"
	"unicode/utf16"
	"unsafe"
)

/*
#include "windowing.h"
#include <windows.h>
#include <windowsx.h>

float get_dpi(HWND hwnd) {
	return GetDpiForWindow(hwnd);
}
*/
import "C"

func asEventType(msg uint32) eventType {
	switch msg {
	case 0x0002:
		fallthrough
	case 0x0012:
		return evtQuit
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

func scaleScrollDelta(delta float32) float32 {
	return delta / 120.0
}

func createWindow(windowName string, width, height int, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	C.window_main(title, C.int(width), C.int(height), evtSharedMem.AsPointer(), evtSharedMemSize)
}

func (w *Window) showWindow(evtSharedMem *evtMem) {
	C.window_show(w.handle)
}

func (w *Window) destroy() {
	C.window_destroy(w.handle)
}

func (w *Window) poll() {
	evtType := uint32(C.window_poll_controller(w.handle))
	if evtType != 0 {
		w.processControllerEvent(asEventType(evtType))
	}
	evtType = 1
	for evtType != 0 && !w.evtSharedMem.IsQuit() {
		evtType = uint32(C.window_poll(w.handle))
		t := asEventType(evtType)
		if w.evtSharedMem.IsResize() {
			t = evtResize
			w.evtSharedMem.ResetHeader()
		}
		if t != evtUnknown {
			w.processEvent(t)
		}
	}
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

func (w *Window) getDPI() (int, int, error) {
	dpi := C.get_dpi(C.HWND(w.handle))
	return int(dpi), int(dpi), nil
}

func (w *Window) openFile(search ...FileSearch) (string, bool) {
	sb := strings.Builder{}
	for _, s := range search {
		sb.WriteString(s.Title)
		sb.WriteString(" (.")
		sb.WriteString(s.Extension)
		sb.WriteString(")\x00*.")
		sb.WriteString(s.Extension)
		sb.WriteString("\x00")
	}
	sb.WriteString("\x00")
	outStr := (*C.char)(C.malloc(0))
	wStr := utf16.Encode([]rune(sb.String()))
	ok := C.window_open_file(w.handle, C.LPCWSTR(unsafe.Pointer(&wStr[0])), &outStr)
	out := C.GoString(outStr)
	C.free(unsafe.Pointer(outStr))
	if ok {
		return out, true
	}
	return "", false
}

func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }
