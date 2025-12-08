//go:build darwin && !ios

package windowing

import (
	"kaiju/klib"
	"unsafe"
)

const macOSSupportIssueID = 485

// Lifecycle and eventing
func (w *Window) createWindow(windowName string, x, y int, _ any) { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) showWindow()                                     { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) poll()                                           { klib.NotYetImplemented(macOSSupportIssueID) }

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

// Destroy expects native handle (window.go calls destroyWindow(w.handle))
func destroyWindow(handle unsafe.Pointer) { klib.NotYetImplemented(macOSSupportIssueID) }

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
	klib.NotYetImplemented(macOSSupportIssueID)
	return 0, 0, nil
}
func (w *Window) screenSizeMM() (int, int, error) {
	klib.NotYetImplemented(macOSSupportIssueID)
	return 0, 0, nil
}
func (w *Window) dotsPerMillimeter() float64 {
	klib.NotYetImplemented(macOSSupportIssueID)
	return 0
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

// Scale mouse wheel delta on macOS
func scaleScrollDelta(delta float32) float32 { return delta }
