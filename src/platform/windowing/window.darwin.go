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
	"kaiju/klib"
	"unsafe"
)

const macOSSupportIssueID = 485

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
func (w *Window) cursorStandard() { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) cursorIbeam()    { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) cursorSizeAll()  { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) cursorSizeNS()   { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) cursorSizeWE()   { klib.NotYetImplemented(macOSSupportIssueID) }

// Clipboard (private)
func (w *Window) copyToClipboard(text string) { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) clipboardContents() string {
	klib.NotYetImplemented(macOSSupportIssueID)
	return ""
}

func destroyWindow(handle unsafe.Pointer) {
	klib.NotYetImplemented(macOSSupportIssueID)
}

// Focus (private)
func (w *Window) focus() { klib.NotYetImplemented(macOSSupportIssueID) }

// Position/Size (private)
func (w *Window) setPosition(x, y int)      { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) setSize(width, height int) { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) position() (x, y int) {
	klib.NotYetImplemented(macOSSupportIssueID)
	return 0, 0
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
func (w *Window) removeBorder() { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) addBorder()    { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) showCursor()   { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) hideCursor()   { klib.NotYetImplemented(macOSSupportIssueID) }

// Cursor lock (private) — window.go passes (x, y)
func (w *Window) lockCursor(x, y int) { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) unlockCursor()       { klib.NotYetImplemented(macOSSupportIssueID) }

// Fullscreen/windowed (private) — window.go calls setWindowed(width, height)
func (w *Window) setFullscreen()                { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) setWindowed(width, height int) { klib.NotYetImplemented(macOSSupportIssueID) }

// Raw mouse input (private)
func (w *Window) disableRawMouse() { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) enableRawMouse()  { klib.NotYetImplemented(macOSSupportIssueID) }

// Title (private)
func (w *Window) setTitle(title string) { klib.NotYetImplemented(macOSSupportIssueID) }

// App asset read (private)
func (w *Window) readApplicationAsset(name string) ([]byte, error) {
	klib.NotYetImplemented(macOSSupportIssueID)
	return nil, nil
}

// cHandle/cInstance used by PlatformWindow/PlatformInstance
func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }
