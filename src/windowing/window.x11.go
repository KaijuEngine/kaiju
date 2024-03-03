//go:build linux || darwin

/******************************************************************************/
/* window.x11.go                                                             */
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

/*
#cgo noescape window_main
#cgo noescape window_show
#cgo noescape window_destroy
#cgo noescape window_focus

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

func createWindow(windowName string, width, height, x, y int, evtSharedMem *evtMem) {
	title := C.CString(string(windowName))
	defer C.free(unsafe.Pointer(title))
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

func (w *Window) cursorSizeAll() {
	klib.NotYetImplemented(258)
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

func (w *Window) focus() {
	C.window_focus(w.handle)
}

func (w *Window) position() (x int, y int) {
	klib.NotYetImplemented(222)
	return -1, -1
}

func (w *Window) setPosition(x, y int) {
	klib.NotYetImplemented(233)
}

func (w *Window) setSize(width, height int) {
	klib.NotYetImplemented(236)
}

func (w *Window) removeBorder() {
	klib.NotYetImplemented(234)
}

func (w *Window) addBorder() {
	klib.NotYetImplemented(234)
}
