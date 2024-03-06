//go:build windows

/******************************************************************************/
/* window.win32.go                                                            */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package windowing

import (
	"kaiju/klib"
	"unicode/utf16"
	"unsafe"
)

/*
#cgo noescape window_main
#cgo noescape window_show
#cgo noescape window_destroy
#cgo noescape window_cursor_standard
#cgo noescape window_cursor_ibeam
#cgo noescape get_dpi
#cgo noescape window_focus
#cgo noescape window_position
#cgo noescape set_window_position
#cgo noescape set_window_size
#cgo noescape remove_border
#cgo noescape add_border

#include "windowing.h"
*/
import "C"

func asEventType(msg uint32) eventType {
	switch msg {
	case 0x0002:
		fallthrough
	case 0x0012:
		return evtQuit
	case 0x0003:
		return evtMove
	case 0x0005:
		return evtResize
	case 0x0006:
		return evtActivity
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

func createWindow(windowName string, width, height, x, y int, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	C.window_main(title, C.int(width), C.int(height),
		C.int(x), C.int(y), evtSharedMem.AsPointer(), evtSharedMemSize)
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
		} else if w.evtSharedMem.IsMove() {
			t = evtMove
			w.evtSharedMem.ResetHeader()
		} else if w.evtSharedMem.IsActivity() {
			t = evtActivity
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

func (w *Window) cursorSizeAll() {
	C.window_cursor_size_all(w.handle)
}

func (w *Window) cursorSizeNS() {
	C.window_cursor_size_ns(w.handle)
}

func (w *Window) cursorSizeWE() {
	C.window_cursor_size_we(w.handle)
}

func (w *Window) copyToClipboard(text string) {
	klib.NotYetImplemented(102)
}

func (w *Window) clipboardContents() string {
	klib.NotYetImplemented(102)
	return ""
}

func (w *Window) getDPI() (int, int, error) {
	dpi := C.get_dpi(w.handle)
	return int(dpi), int(dpi), nil
}

func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

func (w *Window) focus() {
	C.window_focus(w.handle)
}

func (w *Window) position() (x, y int) {
	C.window_position(w.handle, (*C.int)(unsafe.Pointer(&x)), (*C.int)(unsafe.Pointer(&y)))
	return x, y
}

func (w *Window) setPosition(x, y int) {
	C.set_window_position(w.handle, C.int(x), C.int(y))
}

func (w *Window) setSize(width, height int) {
	C.set_window_size(w.handle, C.int(width), C.int(height))
}

func (w *Window) removeBorder() {
	C.remove_border(w.handle)
}

func (w *Window) addBorder() {
	C.add_border(w.handle)
}
