//go:build android

/******************************************************************************/
/* window.win32.go                                                            */
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
	"kaiju/platform/profiler/tracing"
	"unsafe"
)

/*
#cgo noescape window_main
#cgo noescape window_poll
#cgo noescape pull_android_window
#include <stdint.h>
#include "windowing.h"
*/
import "C"

//export goProcessEvents
func goProcessEvents(goWindow C.uint64_t, events unsafe.Pointer, eventCount C.uint32_t) {
	goProcessEventsCommon(uint64(goWindow), events, uint32(eventCount))
}

func scaleScrollDelta(delta float32) float32 {
	return 0
}

func (w *Window) createWindow(_ string, _, _ int, platformState any) {
	w.handle = platformState.(unsafe.Pointer)
	w.lookupId = nextLookupId.Add(1)
	windowLookup.Store(w.lookupId, w)
	C.window_main(w.handle, C.uint64_t(w.lookupId))
	w.instance = unsafe.Pointer(C.pull_android_window(w.handle))
}

func (w *Window) showWindow() {}

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

func (w *Window) cursorStandard() {
}

func (w *Window) cursorIbeam() {
}

func (w *Window) cursorSizeAll() {
}

func (w *Window) cursorSizeNS() {
}

func (w *Window) cursorSizeWE() {
}

func (w *Window) copyToClipboard(text string) {
}

func (w *Window) clipboardContents() string {
	return ""
}

func (w *Window) dotsPerMillimeter() float64 {
	return 0
}

func (w *Window) sizeMM() (int, int, error) {
	return 0, 0, nil
}

func (w *Window) screenSizeMM() (int, int, error) {
	return 0, 0, nil
}

func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

func (w *Window) focus() {
}

func (w *Window) position() (x, y int) {
	return 0, 0
}

func (w *Window) setPosition(x, y int) {
}

func (w *Window) setSize(width, height int) {
}

func (w *Window) removeBorder() {
}

func (w *Window) addBorder() {
}

func (w *Window) showCursor() {
}

func (w *Window) hideCursor() {
}

func (w *Window) lockCursor(x, y int) {
}

func (w *Window) unlockCursor() {
}

func (w *Window) setFullscreen() {
}

func (w *Window) setWindowed(width, height int) {
}

func (w *Window) enableRawMouse() {
}

func (w *Window) disableRawMouse() {
}

func (w Window) setTitle(newTitle string) {
}
