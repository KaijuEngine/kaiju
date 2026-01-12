//go:build linux && !android && !wayland

/******************************************************************************/
/* window.linux.go                                                            */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
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

/*
#cgo CFLAGS: -I/usr/include
#cgo LDFLAGS: -lX11 -lXcursor
#cgo noescape get_toggle_key_state
#cgo noescape window_main
#cgo noescape window_poll_controller
#cgo noescape window_poll
#cgo noescape window_show
#cgo noescape window_destroy
#cgo noescape window_focus
#cgo noescape window_position
#cgo noescape window_set_position
#cgo noescape window_set_size
#cgo noescape window_width_mm
#cgo noescape window_height_mm
#cgo noescape window_cursor_standard
#cgo noescape window_cursor_ibeam
#cgo noescape window_cursor_size_all
#cgo noescape window_cursor_size_ns
#cgo noescape window_cursor_size_we
#cgo noescape window_show_cursor
#cgo noescape window_hide_cursor
#cgo noescape window_dpi
#cgo noescape window_set_title
#cgo noescape window_set_full_screen
#cgo noescape window_set_windowed
#cgo noescape window_lock_cursor
#cgo noescape window_unlock_cursor

#include <stdlib.h>
#include "windowing.h"
*/
import "C"
import (
	"errors"
	"kaiju/klib"
	"kaiju/platform/hid"
	"unsafe"

	"golang.design/x/clipboard"
)

//export goProcessEvents
func goProcessEvents(goWindow C.uint64_t, events unsafe.Pointer, eventCount C.uint32_t) {
	goProcessEventsCommon(uint64(goWindow), events, uint32(eventCount))
}

func (w *Window) checkToggleKeyState() map[hid.KeyboardKey]bool {
	mask := C.get_toggle_key_state()

	caps := (mask & 1) != 0
	num := (mask & 2) != 0
	scroll := (mask & 4) != 0

	return map[hid.KeyboardKey]bool{
		hid.KeyboardKeyCapsLock:   caps,
		hid.KeyboardKeyNumLock:    num,
		hid.KeyboardKeyScrollLock: scroll,
	}
}

func (w *Window) createWindow(windowName string, x, y int, _ any) {
	title := C.CString(windowName)
	defer C.free(unsafe.Pointer(title))
	w.lookupId = nextLookupId.Add(1)
	windowLookup.Store(w.lookupId, w)
	C.window_main(title, C.int(w.width), C.int(w.height),
		C.int(x), C.int(y), C.uint64_t(w.lookupId))
}

func (w *Window) showWindow() {
	C.window_show(w.handle)
}

func destroyWindow(handle unsafe.Pointer) {
	C.window_destroy(handle)
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
	width := C.window_width_mm(w.handle)
	height := C.window_height_mm(w.handle)
	return int(width), int(height), nil
}

func (w *Window) cHandle() unsafe.Pointer   { return C.window(w.handle) }
func (w *Window) cInstance() unsafe.Pointer { return C.display(w.handle) }

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

func (w *Window) showCursor() {
	C.window_show_cursor(w.handle)
}

func (w *Window) hideCursor() {
	C.window_hide_cursor(w.handle)
}

func (w *Window) dotsPerMillimeter() float64 {
	return float64(C.window_dpi(w.handle))
}

func (w *Window) screenSizeMM() (int, int, error) {
	mm := float64(C.window_dpi(w.handle))
	return int(float64(w.width) * mm), int(float64(w.height) * mm), nil
}

func (w Window) setTitle(name string) {
	title := C.CString(name)
	defer C.free(unsafe.Pointer(title))
	C.window_set_title(w.handle, title)
}

func (w Window) setFullscreen() {
	C.window_set_full_screen(w.handle)
}

func (w Window) setWindowed(width, height int) {
	C.window_set_windowed(w.handle, C.int(width), C.int(height))
}

func (w Window) lockCursor(x, y int) {
	C.window_lock_cursor(w.handle, C.int(x), C.int(y))
}

func (w Window) unlockCursor() {
	C.window_unlock_cursor(w.handle)
}

func (w *Window) removeBorder() {
	klib.NotYetImplemented(234)
}

func (w *Window) addBorder() {
	klib.NotYetImplemented(234)
}

func (w Window) disableRawMouse() { /* Don't think this is needed for X11 */ }
func (w Window) enableRawMouse()  { /* Don't think this is needed for X11 */ }

func (w *Window) readApplicationAsset(path string) ([]byte, error) {
	return []byte{}, errors.New("linux doesn't support application assets")
}
