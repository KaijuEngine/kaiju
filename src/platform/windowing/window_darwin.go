//go:build darwin && !ios

/******************************************************************************/
/* window_darwin.go                                                           */
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

import (
	"kaiju/klib"
	"kaiju/platform/hid"

	"unsafe"
)

const macOSSupportIssueID = 485

func (w *Window) checkToggleKeyState() map[hid.KeyboardKey]bool {
	klib.NotYetImplemented(494)
	return map[hid.KeyboardKey]bool{}
}

// Lifecycle and eventing
func (w *Window) createWindow(windowName string, x, y int, _ any) {
	klib.NotYetImplemented(macOSSupportIssueID)
}
func (w *Window) showWindow() { klib.NotYetImplemented(macOSSupportIssueID) }
func (w *Window) poll()       { klib.NotYetImplemented(macOSSupportIssueID) }

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
