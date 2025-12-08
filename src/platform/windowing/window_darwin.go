//go:build darwin && !ios

package windowing

import "unsafe"

// Lifecycle and eventing
func (w *Window) createWindow(windowName string, x, y int, _ any) {}
func (w *Window) showWindow()                                     {}
func (w *Window) poll()                                           {}

// Cursor variants (private)
func (w *Window) cursorStandard() {}
func (w *Window) cursorIbeam()    {}
func (w *Window) cursorSizeAll()  {}
func (w *Window) cursorSizeNS()   {}
func (w *Window) cursorSizeWE()   {}

// Clipboard (private)
func (w *Window) copyToClipboard(text string) {}
func (w *Window) clipboardContents() string   { return "" }

// Destroy expects native handle (window.go calls destroyWindow(w.handle))
func destroyWindow(handle unsafe.Pointer) {}

// Focus (private)
func (w *Window) focus() {}

// Position/Size (private)
func (w *Window) setPosition(x, y int)      {}
func (w *Window) setSize(width, height int) {}
func (w *Window) position() (x, y int)      { return 0, 0 }

// Physical metrics (private)
func (w *Window) sizeMM() (int, int, error)       { return 0, 0, nil }
func (w *Window) screenSizeMM() (int, int, error) { return 0, 0, nil }
func (w *Window) dotsPerMillimeter() float64      { return 0 }

// Window decoration and cursor visibility (private)
func (w *Window) removeBorder() {}
func (w *Window) addBorder()    {}
func (w *Window) showCursor()   {}
func (w *Window) hideCursor()   {}

// Cursor lock (private) — window.go passes (x, y)
func (w *Window) lockCursor(x, y int) {}
func (w *Window) unlockCursor()       {}

// Fullscreen/windowed (private) — window.go calls setWindowed(width, height)
func (w *Window) setFullscreen()                {}
func (w *Window) setWindowed(width, height int) {}

// Raw mouse input (private)
func (w *Window) disableRawMouse() {}
func (w *Window) enableRawMouse()  {}

// Title (private)
func (w *Window) setTitle(title string) {}

// App asset read (private)
func (w *Window) readApplicationAsset(name string) ([]byte, error) {
	return nil, nil
}

// cHandle/cInstance used by PlatformWindow/PlatformInstance
func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

// Scale mouse wheel delta on macOS (stub: passthrough; adjust later if needed)
func scaleScrollDelta(delta float32) float32 { return delta }
