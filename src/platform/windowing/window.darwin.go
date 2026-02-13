//go:build darwin && !ios

package windowing

/******************************************************************************/
/* window.darwin.go                                                           */
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

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore -framework Metal
#include "cocoa_window.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"image"
	"os"
	"unsafe"

	"kaiju/klib"
	"kaiju/platform/hid"
)

//export goProcessEvents
func goProcessEvents(goWindow C.uint64_t, events unsafe.Pointer, eventCount C.uint32_t) {
	goProcessEventsCommon(uint64(goWindow), events, uint32(eventCount))
}

// Scroll delta is already scaled to match Windows in cocoa_window.m (* 120),
// so just pass through without additional scaling
func scaleScrollDelta(delta float32) float32 {
	return delta
}

func (w *Window) checkToggleKeyState() map[hid.KeyboardKey]bool {
	// macOS only has Caps Lock as a toggle key
	caps := bool(C.cocoa_get_caps_lock_toggle_key_state())

	return map[hid.KeyboardKey]bool{
		hid.KeyboardKeyCapsLock: caps,
	}
}

// Lifecycle and eventing
func (w *Window) createWindow(windowName string, x, y int, _ any) {
	cTitle := C.CString(windowName)
	defer C.free(unsafe.Pointer(cTitle))

	w.lookupId = nextLookupId.Add(1)
	windowLookup.Store(w.lookupId, w)

	// cocoa_create_window sends WINDOW_EVENT_TYPE_SET_HANDLE to set w.handle and w.instance
	var nsWindow unsafe.Pointer
	C.cocoa_create_window(cTitle, C.int(x), C.int(y), C.int(w.width), C.int(w.height), &nsWindow, unsafe.Pointer(uintptr(w.lookupId)))
}

func (w *Window) showWindow() {
	if w.instance != nil {
		C.cocoa_show_window(w.instance)
	}
}

func (w *Window) poll() {
	if w.instance != nil {
		C.cocoa_poll_events(w.instance)
	}
}

// Cursor variants (private)
func (w *Window) cursorStandard() { C.cocoa_cursor_standard() }
func (w *Window) cursorIbeam()    { C.cocoa_cursor_ibeam() }
func (w *Window) cursorSizeAll()  { C.cocoa_cursor_size_all() }
func (w *Window) cursorSizeNS()   { C.cocoa_cursor_size_ns() }
func (w *Window) cursorSizeWE()   { C.cocoa_cursor_size_we() }

// Clipboard (private)
func (w *Window) copyToClipboard(text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.cocoa_copy_to_clipboard(cText)
}

func (w *Window) clipboardContents() string {
	cStr := C.cocoa_clipboard_contents()
	if cStr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cStr))
	return C.GoString(cStr)
}

func destroyWindow(instance unsafe.Pointer) {
	if instance != nil {
		C.cocoa_destroy_window(instance)
	}
}

// Focus (private)
func (w *Window) focus() {
	if w.instance != nil {
		C.cocoa_focus_window(w.instance)
	}
}

// Position/Size (private)
func (w *Window) setPosition(x, y int) {
	if w.instance != nil {
		C.cocoa_set_position(w.instance, C.int(x), C.int(y))
	}
}

func (w *Window) setSize(width, height int) {
	if w.instance != nil {
		C.cocoa_set_size(w.instance, C.int(width), C.int(height))
	}
}

func (w *Window) position() (x, y int) {
	if w.instance == nil {
		return 0, 0
	}
	var cx, cy C.int
	C.cocoa_get_position(w.instance, &cx, &cy)
	return int(cx), int(cy)
}

// Physical metrics
func (w *Window) sizeMM() (int, int, error) {
	dpm := w.dotsPerMillimeter()
	if dpm <= 0 {
		return 0, 0, nil
	}

	return int(float64(w.width) / dpm),
		int(float64(w.height) / dpm),
		nil
}

func (w *Window) screenSizeMM() (int, int, error) {
	dpm := w.dotsPerMillimeter()
	if dpm <= 0 {
		return 0, 0, nil
	}

	pw := float64(C.cocoa_get_screen_pixel_width(w.instance))
	ph := float64(C.cocoa_get_screen_pixel_height(w.instance))

	return int(pw / dpm),
		int(ph / dpm),
		nil
}

func (w *Window) dotsPerMillimeter() float64 {
	if w.instance == nil {
		return 0
	}

	scale := float64(C.cocoa_get_backing_scale_factor(w.instance))
	if scale <= 0 {
		scale = 1
	}

	// macOS logical DPI model
	dpi := 96.0 * scale
	return dpi / 25.4
}

// Window decoration and cursor visibility (private)
func (w *Window) removeBorder() {
	if w.instance != nil {
		C.cocoa_remove_border(w.instance)
	}
}

func (w *Window) addBorder() {
	if w.instance != nil {
		C.cocoa_add_border(w.instance)
	}
}

func (w *Window) showCursor() { C.cocoa_show_cursor() }
func (w *Window) hideCursor() { C.cocoa_hide_cursor() }

// Cursor lock (private) — window.go passes (x, y)
func (w *Window) lockCursor(x, y int) {
	if w.instance != nil {
		C.cocoa_lock_cursor(w.instance, C.int(x), C.int(y))
	}
}

func (w *Window) unlockCursor() {
	if w.instance != nil {
		C.cocoa_unlock_cursor(w.instance)
	}
}

// Fullscreen/windowed (private) — window.go calls setWindowed(width, height)
func (w *Window) setFullscreen() {
	if w.instance != nil {
		C.cocoa_set_fullscreen(w.instance)
	}
}

func (w *Window) setWindowed(width, height int) {
	if w.instance != nil {
		C.cocoa_set_windowed(w.instance, C.int(width), C.int(height))
	}
}

// Raw mouse input (private)
func (w *Window) disableRawMouse() {
	if w.instance != nil {
		C.cocoa_disable_raw_mouse(w.instance)
	}
}

func (w *Window) enableRawMouse() {
	if w.instance != nil {
		C.cocoa_enable_raw_mouse(w.instance)
	}
}

// Title (private)
func (w *Window) setTitle(title string) {
	if w.instance != nil {
		cTitle := C.CString(title)
		defer C.free(unsafe.Pointer(cTitle))
		C.cocoa_set_title(w.instance, cTitle)
	}
}

func (w *Window) setCursorPosition(x, y int) {
	// C.cocoa_set_cursor_position(w.handle, C.int(x), C.int(y))
}

func (w *Window) setIcon(img image.Image) {
	klib.NotYetImplemented(626)
}

// App asset read (private)
// On macOS, application assets are stored in the app bundle's Resources folder.
// This function reads files from Contents/Resources/ in the .app bundle.
func (w *Window) readApplicationAsset(name string) ([]byte, error) {
	// Try to get the main bundle's resource path
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	// Use Cocoa to get the bundle resource path
	var bundlePath unsafe.Pointer
	C.cocoa_get_bundle_resource_path(cName, &bundlePath)

	if bundlePath == nil {
		// Not in a bundle or resource not found, try direct file read
		data, err := os.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("asset not found in bundle or filesystem: %s", name)
		}
		return data, nil
	}

	goPath := C.GoString((*C.char)(bundlePath))
	C.free(bundlePath)

	return os.ReadFile(goPath)
}

// cHandle/cInstance used by PlatformWindow/PlatformInstance
func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

func CocoaRunApp() {
	C.cocoa_run_app()
}
