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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package windowing

import (
	"unicode/utf16"
	"unsafe"

	"golang.design/x/clipboard"
)

/*
#cgo LDFLAGS: -lgdi32 -lXInput
#cgo noescape window_main
#cgo noescape window_show
#cgo noescape window_destroy
#cgo noescape window_cursor_standard
#cgo noescape window_cursor_ibeam
#cgo noescape window_dpi
#cgo noescape window_focus
#cgo noescape window_position
#cgo noescape window_set_position
#cgo noescape window_set_size
#cgo noescape window_remove_border
#cgo noescape window_add_border
#cgo noescape window_poll_controller
#cgo noescape window_poll

#include "windowing.h"
*/
import "C"

func scaleScrollDelta(delta float32) float32 {
	return delta / 120.0
}

func (w *Window) createWindow(windowName string, win *Window, x, y int) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	goWindow := uint64(uintptr(unsafe.Pointer(win)))
	C.window_main(title, C.int(win.width), C.int(win.height),
		C.int(x), C.int(y), C.uint64_t(goWindow))
}

func (w *Window) showWindow() {
	C.window_show(w.handle)
}

func (w *Window) destroy() {
	C.window_destroy(w.handle)
}

//export goProcessEvents
func goProcessEvents(goWindow uint64, events unsafe.Pointer, eventCount uint32) {
	var win *Window
	gw := unsafe.Pointer(uintptr(goWindow))
	for i := range activeWindows {
		if unsafe.Pointer(activeWindows[i]) == gw {
			win = activeWindows[i]
			break
		}
	}
	for range eventCount {
		eType, body := readType(events)
		switch eType {
		case windowEventTypeSetHandle:
			evt := asSetHandleEvent(body)
			win.handle = evt.hwnd
			win.instance = evt.instance
		case windowEventTypeActivity:
			win.processWindowActivityEvent(asWindowActivityEvent(body))
		case windowEventTypeMove:
			win.processWindowMoveEvent(asWindowMoveEvent(body))
		case windowEventTypeResize:
			win.processWindowResizeEvent(asWindowResizeEvent(body))
		case windowEventTypeMouseMove:
			win.processMouseMoveEvent(asMouseMoveWindowEvent(body))
		case windowEventTypeMouseScroll:
			win.processMouseScrollEvent(asMouseScrollWindowEvent(body))
		case windowEventTypeMouseButton:
			win.processMouseButtonEvent(asMouseButtonWindowEvent(body))
		case windowEventTypeKeyboardButton:
			win.processKeyboardButtonEvent(asKeyboardButtonWindowEvent(body))
		case windowEventTypeControllerState:
			win.processControllerStateEvent(asControllerStateWindowEvent(body))
		case windowEventTypeFatal:
			events = body
			win.fatalFromNativeAPI = true
			break
		}
		events = unsafe.Pointer(uintptr(body) + evtUnionSize)
	}
}

func (w *Window) poll() {
	C.window_poll_controller(w.handle)
	C.window_poll(w.handle)
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
	clipboard.Write(clipboard.FmtText, []byte(text))
}

func (w *Window) clipboardContents() string {
	return string(clipboard.Read(clipboard.FmtText))
}

func (w *Window) sizeMM() (int, int, error) {
	dpi := float64(C.window_dpi(w.handle))
	mm := dpi / 25.4
	return int(float64(w.width) * mm), int(float64(w.height) * mm), nil
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
	C.window_set_position(w.handle, C.int(x), C.int(y))
}

func (w *Window) setSize(width, height int) {
	C.window_set_size(w.handle, C.int(width), C.int(height))
}

func (w *Window) removeBorder() {
	C.window_remove_border(w.handle)
}

func (w *Window) addBorder() {
	C.window_add_border(w.handle)
}
