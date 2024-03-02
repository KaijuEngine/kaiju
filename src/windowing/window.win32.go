//go:build windows

/******************************************************************************/
/* window.win32.go                                                           */
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
#cgo noescape window_destroy
#cgo noescape get_dpi

#include <assert.h>
#include "windowing.h"

// Force the compiler not to strip functions called from assembly
void forceLink() {
	assert(&window_show != NULL);
	assert(&window_cursor_standard != NULL);
	assert(&window_cursor_ibeam != NULL);
	assert(&window_focus != NULL);
	assert(&window_position != NULL);
	assert(&set_window_position != NULL);
	assert(&set_window_size != NULL);
	assert(&remove_border != NULL);
	assert(&add_border != NULL);
}
*/
import "C"

func init() { C.forceLink() }

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
	cShowWindow(w.handle)
}

func (w *Window) destroy() {
	C.window_destroy(w.handle)
}

func (w *Window) poll() {
	evtType := uint32(cWindowPollController(w.handle))
	if evtType != 0 {
		w.processControllerEvent(asEventType(evtType))
	}
	evtType = 1
	for evtType != 0 && !w.evtSharedMem.IsQuit() {
		evtType = uint32(cWindowPoll(w.handle))
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
	cWindowCursorStandard(w.handle)
}

func (w *Window) cursorIbeam() {
	cWindowCursorIbeam(w.handle)
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
	cWindowFocus(w.handle)
}

func (w *Window) position() (int, int) {
	var px, py int32
	cWindowPosition(w.handle, &px, &py)
	return int(px), int(py)
}

func (w *Window) setPosition(x, y int) {
	cSetWindowPosition(w.handle, int32(x), int32(y))
}

func (w *Window) setSize(width, height int) {
	cSetWindowSize(w.handle, int32(width), int32(height))
}

func (w *Window) removeBorder() {
	cRemoveBorder(w.handle)
}

func (w *Window) addBorder() {
	cAddBorder(w.handle)
}
