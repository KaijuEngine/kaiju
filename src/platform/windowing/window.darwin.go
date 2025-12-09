//go:build darwin && !ios

package windowing

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework QuartzCore -framework Metal
#include "cocoa_window.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
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

func destroyWindow(handle unsafe.Pointer) {
	if handle != nil {
		C.cocoa_destroy_window(handle)
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

// Physical metrics (private)
func (w *Window) sizeMM() (int, int, error) {
	if w.instance == nil {
		return 0, 0, nil
	}
	// Convert DPI to dots per millimeter, then multiply by window dimensions
	dpi := float64(C.cocoa_get_dpi(w.instance))
	mm := dpi / 25.4 // 1 inch = 25.4mm
	return int(float64(w.width) * mm), int(float64(w.height) * mm), nil
}

func (w *Window) screenSizeMM() (int, int, error) {
	if w.instance == nil {
		return 0, 0, nil
	}
	var width, height C.int
	C.cocoa_screen_size_mm(w.instance, &width, &height)
	return int(width), int(height), nil
}

func (w *Window) dotsPerMillimeter() float64 {
	if w.instance == nil {
		return 0
	}
	dpi := float64(C.cocoa_get_dpi(w.instance))
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
