//go:build android

/******************************************************************************/
/* window.android.go                                                          */
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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package windowing

import (
	"errors"
	"unsafe"
)

func asEventType(msg int, e *evtMem) eventType {
	switch msg {
	// TODO:  Fill out the other cases
	default:
		return evtUnknown
	}
}

func scaleScrollDelta(delta float32) float32 { return delta }

func createWindow(windowName string, width, height, x, y int, evtSharedMem *evtMem) {
	// TODO:  implement
}

func (w *Window) showWindow(evtSharedMem *evtMem) {
	// TODO:  implement
}

func (w *Window) destroy() {
	// TODO:  implement
}

func (w *Window) poll() {
	// TODO:  implement
}

func (w *Window) cursorStandard() {}
func (w *Window) cursorIbeam()    {}
func (w *Window) cursorSizeAll()  {}
func (w *Window) cursorSizeNS()   {}
func (w *Window) cursorSizeWE()   {}

func (w *Window) copyToClipboard(text string) {
	// TODO:  implement
}

func (w *Window) clipboardContents() string {
	// TODO:  implement
	return ""
}

func (w *Window) sizeMM() (int, int, error) {
	// TODO:  implement
	return 0, 0, errors.New("not implemented")
}

func (w *Window) cHandle() unsafe.Pointer {
	// TODO:  implement
	return nil
}

func (w *Window) cInstance() unsafe.Pointer {
	// TODO:  implement
	return nil
}

func (w *Window) focus()                    {}
func (w *Window) position() (x, y int)      { return 0, 0 }
func (w *Window) setPosition(x, y int)      {}
func (w *Window) setSize(width, height int) {}
func (w *Window) removeBorder()             {}
func (w *Window) addBorder()                {}
