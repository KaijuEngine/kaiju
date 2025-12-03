//go:build android

/******************************************************************************/
/* window_android.go                                                          */
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
	"fmt"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"unsafe"
)

/*
#cgo noescape window_main
#cgo nocallback window_main
#cgo noescape window_poll
#cgo noescape pull_android_window
#cgo nocallback pull_android_window
#cgo noescape window_size_mm
#cgo nocallback window_size_mm
#cgo noescape window_open_website
#cgo nocallback window_asset_exists
#cgo noescape window_asset_exists
#cgo nocallback window_asset_length
#cgo noescape window_asset_length
#cgo nocallback window_asset_read
#cgo noescape window_asset_read
#include <stdint.h>
#include <stdlib.h>
#include "windowing.h"
*/
import "C"

//export goProcessEvents
func goProcessEvents(goWindow C.uint64_t, events unsafe.Pointer, eventCount C.uint32_t) {
	goProcessEventsCommon(uint64(goWindow), events, uint32(eventCount))
}

func scaleScrollDelta(delta float32) float32 {
	return 1
}

func (w *Window) createWindow(_ string, _, _ int, platformState any) {
	w.handle = platformState.(unsafe.Pointer)
	w.lookupId = nextLookupId.Add(1)
	windowLookup.Store(w.lookupId, w)
	C.window_main(w.handle, C.uint64_t(w.lookupId))
	w.instance = unsafe.Pointer(C.pull_android_window(w.handle))
	klib.OpenWebsiteAndroidFunc = w.openWebsite
}

func destroyWindow(handle unsafe.Pointer) {
}

func (w *Window) pollController() {
}

func (w *Window) pollEvents() {
	defer tracing.NewRegion("Window.pollEvents").End()
	C.window_poll(w.handle)
}

func (w *Window) poll() {
	w.pollController()
	w.pollEvents()
}

func (w *Window) copyToClipboard(text string) {
}

func (w *Window) clipboardContents() string {
	return ""
}

func (w *Window) dotsPerMillimeter() float64 {
	wmm, hmm, err := w.screenSizeMM()
	if err != nil || wmm == 0 || hmm == 0 {
		return 1.0
	}
	dpmW := float64(w.width) / float64(wmm)
	dpmH := float64(w.height) / float64(hmm)
	return (dpmW + dpmH) / 2.0
}

func (w *Window) sizeMM() (int, int, error) {
	return w.screenSizeMM()
}

func (w *Window) screenSizeMM() (int, int, error) {
	var wmm, hmm int
	C.window_size_mm(w.handle, (*C.int)(unsafe.Pointer(&wmm)), (*C.int)(unsafe.Pointer(&hmm)))
	return wmm, hmm, nil
}

func (w *Window) openWebsite(url string) {
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))
	C.window_open_website(w.handle, cURL)
}

func (w *Window) readApplicationAsset(path string) ([]byte, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	if !C.window_asset_exists(w.handle, cPath) {
		return []byte{}, fmt.Errorf("application asset file '%s' doesn't exist", path)
	}
	assetLen := int64(C.window_asset_length(w.handle, cPath))
	if assetLen <= 0 {
		return []byte{}, fmt.Errorf("failed to read the asset length for '%s'", path)
	}
	buff := make([]byte, assetLen)
	outLen := int64(C.window_asset_read(w.handle, cPath, unsafe.Pointer(&buff[0])))
	if outLen == 0 {
		return []byte{}, fmt.Errorf("failed to read the data in asset '%s'", path)
	}
	return buff, nil
}

func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

func (w *Window) showWindow()                   {}
func (w *Window) cursorStandard()               {}
func (w *Window) cursorIbeam()                  {}
func (w *Window) cursorSizeAll()                {}
func (w *Window) cursorSizeNS()                 {}
func (w *Window) cursorSizeWE()                 {}
func (w *Window) focus()                        {}
func (w *Window) position() (x, y int)          { return 0, 0 }
func (w *Window) setPosition(x, y int)          {}
func (w *Window) setSize(width, height int)     {}
func (w *Window) removeBorder()                 {}
func (w *Window) addBorder()                    {}
func (w *Window) showCursor()                   {}
func (w *Window) hideCursor()                   {}
func (w *Window) lockCursor(x, y int)           {}
func (w *Window) unlockCursor()                 {}
func (w *Window) setFullscreen()                {}
func (w *Window) setWindowed(width, height int) {}
func (w *Window) enableRawMouse()               {}
func (w *Window) disableRawMouse()              {}
func (w Window) setTitle(newTitle string)       {}
