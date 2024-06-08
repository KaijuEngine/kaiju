//go:build windows

/******************************************************************************/
/* window.win32.go                                                            */
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
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.design/x/clipboard"
	"golang.org/x/sys/windows"
)

/*
#cgo LDFLAGS: -lgdi32 -lXInput
#cgo noescape window_main
//#cgo noescape window_show
#cgo noescape window_destroy
//#cgo noescape window_cursor_standard
//#cgo noescape window_cursor_ibeam
//#cgo noescape window_dpi
//#cgo noescape window_focus
//#cgo noescape window_position
//#cgo noescape window_set_position
//#cgo noescape window_set_size
//#cgo noescape window_remove_border
//#cgo noescape window_add_border

#include "windowing.h"
*/
import "C"

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procShowWindow         = user32.NewProc("ShowWindow")
	procPostMessageA       = user32.NewProc("PostMessageA")
	procGetDpiForWindow    = user32.NewProc("GetDpiForWindow")
	procBringWindowToTop   = user32.NewProc("BringWindowToTop")
	procSetFocus           = user32.NewProc("SetFocus")
	procGetWindowPlacement = user32.NewProc("GetWindowPlacement")
	procSetWindowPos       = user32.NewProc("SetWindowPos")
	procGetWindowLongW     = user32.NewProc("GetWindowLongW")
	procSetWindowLongW     = user32.NewProc("SetWindowLongW")
)

type POINT struct {
	X int32
	Y int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type WINDOWPLACEMENT struct {
	Length           uint32
	Flags            uint32
	ShowCmd          uint32
	PtMinPosition    POINT
	PtMaxPosition    POINT
	RcNormalPosition RECT
}

func asEventType(msg uint32) eventType {
	switch msg {
	case 0x0002:
		fallthrough
	case 0x0012:
		return evtQuit
	case 0x0003:
		return evtMove
	case 0x0005:
		return evtResize
	case 0x0006:
		return evtActivity
	case 0x0104:
		fallthrough
	case 0x0100:
		return evtKeyDown
	case 0x0105:
		fallthrough
	case 0x0101:
		return evtKeyUp
	case 512:
		return evtMouseMove
	case 513:
		return evtLeftMouseDown
	case 514:
		return evtLeftMouseUp
	case 516:
		return evtRightMouseDown
	case 517:
		return evtRightMouseUp
	case 519:
		return evtMiddleMouseDown
	case 520:
		return evtMiddleMouseUp
	case 523:
		return evtX1MouseDown
	case 524:
		return evtX1MouseUp
	case 0x020A:
		return evtMouseWheelVertical
	case 0x020E:
		return evtMouseWheelHorizontal
	case 0xFFFFFFFF - 1:
		return evtControllerStates
	default:
		return evtUnknown
	}
}

func scaleScrollDelta(delta float32) float32 {
	return delta / 120.0
}

func createWindow(windowName string, width, height, x, y int, evtSharedMem *evtMem) {
	windowTitle := utf16.Encode([]rune(windowName))
	title := (*C.wchar_t)(unsafe.Pointer(&windowTitle[0]))
	C.window_main(title, C.int(width), C.int(height),
		C.int(x), C.int(y), evtSharedMem.AsPointer(), evtSharedMemSize)
}

func (w *Window) showWindow(evtSharedMem *evtMem) {
	syscall.SyscallN(procShowWindow.Addr(), uintptr(w.handle), 5 /*SW_SHOW*/)
	//C.window_show(w.handle)
}

func (w *Window) destroy() {
	C.window_destroy(w.handle)
}

func (w *Window) poll() {
	evtType := uint32(C.window_poll_controller(w.handle))
	if evtType != 0 {
		w.processControllerEvent(asEventType(evtType))
	}
	evtType = 1
	for evtType != 0 && !w.evtSharedMem.IsQuit() {
		evtType = uint32(C.window_poll(w.handle))
		t := asEventType(evtType)
		if w.evtSharedMem.IsResize() {
			t = evtResize
			w.evtSharedMem.ResetHeader()
		} else if w.evtSharedMem.IsMove() {
			t = evtMove
			w.evtSharedMem.ResetHeader()
		} else if w.evtSharedMem.IsActivity() {
			t = evtActivity
			w.evtSharedMem.ResetHeader()
		}
		if t != evtUnknown {
			w.processEvent(t)
		}
	}
}

func (w *Window) cursorStandard() {
	syscall.SyscallN(procPostMessageA.Addr(), uintptr(w.handle),
		0x0400+0x0001 /*UWM_SET_CURSOR*/, 1 /*CURSOR_ARROW*/, 0)
	//C.window_cursor_standard(w.handle)
}

func (w *Window) cursorIbeam() {
	syscall.SyscallN(procPostMessageA.Addr(), uintptr(w.handle),
		0x0400+0x0001 /*UWM_SET_CURSOR*/, 2 /*CURSOR_IBEAM*/, 0)
	//C.window_cursor_ibeam(w.handle)
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
	r1, _, _ := syscall.SyscallN(procGetDpiForWindow.Addr(), uintptr(w.handle))
	dpi := float64(r1)
	//dpi := float64(C.window_dpi(w.handle))
	mm := dpi / 25.4
	return int(float64(w.width) * mm), int(float64(w.height) * mm), nil
}

func (w *Window) cHandle() unsafe.Pointer   { return w.handle }
func (w *Window) cInstance() unsafe.Pointer { return w.instance }

func (w *Window) focus() {
	syscall.SyscallN(procBringWindowToTop.Addr(), uintptr(w.handle))
	syscall.SyscallN(procSetFocus.Addr(), uintptr(w.handle))
	//C.window_focus(w.handle)
}

func (w *Window) position() (x, y int) {
	wp := WINDOWPLACEMENT{}
	wp.Length = uint32(unsafe.Sizeof(wp))
	r1, _, _ := syscall.SyscallN(procGetWindowPlacement.Addr(),
		uintptr(w.handle), uintptr(unsafe.Pointer(&wp)))
	if r1 == 0 {
		return 0, 0
	}
	return int(wp.RcNormalPosition.Left), int(wp.RcNormalPosition.Top)
	//C.window_position(w.handle, (*C.int)(unsafe.Pointer(&x)), (*C.int)(unsafe.Pointer(&y)))
	//return x, y
}

func (w *Window) setPosition(x, y int) {
	syscall.SyscallN(procSetWindowPos.Addr(), uintptr(w.handle),
		0, uintptr(unsafe.Pointer(&x)), uintptr(unsafe.Pointer(&y)),
		0, 0, 0x0001|0x0004 /*SWP_NOSIZE|SWP_NOZORDER*/)
	//C.window_set_position(w.handle, C.int(x), C.int(y))
}

func (w *Window) setSize(width, height int) {
	syscall.SyscallN(procSetWindowPos.Addr(), uintptr(w.handle),
		0, 0, 0, uintptr(unsafe.Pointer(&width)),
		uintptr(unsafe.Pointer(&height)), 0x0002|0x0004 /*SWP_NOMOVE|SWP_NOZORDER*/)
	//C.window_set_size(w.handle, C.int(width), C.int(height))
}

func (w *Window) removeBorder() {
	gwlStyle := -16 /*GWL_STYLE*/
	r1, _, _ := syscall.SyscallN(procGetWindowLongW.Addr(),
		uintptr(w.handle), uintptr(gwlStyle))
	style := int32(r1)
	style &= ^0x00C00000 /*WS_CAPTION*/
	style &= ^0x00040000 /*WS_THICKFRAME*/
	style &= ^0x00020000 /*WS_MINIMIZEBOX*/
	style &= ^0x00010000 /*WS_MAXIMIZEBOX*/
	style &= ^0x00080000 /*WS_SYSMENU*/
	syscall.SyscallN(procSetWindowLongW.Addr(),
		uintptr(w.handle), uintptr(gwlStyle), uintptr(style))
	//C.window_remove_border(w.handle)
}

func (w *Window) addBorder() {
	gwlStyle := -16 /*GWL_STYLE*/
	r1, _, _ := syscall.SyscallN(procGetWindowLongW.Addr(),
		uintptr(w.handle), uintptr(gwlStyle))
	style := int32(r1)
	style |= 0x00C00000 /*WS_CAPTION*/
	style |= 0x00040000 /*WS_THICKFRAME*/
	style |= 0x00020000 /*WS_MINIMIZEBOX*/
	style |= 0x00010000 /*WS_MAXIMIZEBOX*/
	style |= 0x00080000 /*WS_SYSMENU*/
	syscall.SyscallN(procSetWindowLongW.Addr(),
		uintptr(w.handle), uintptr(gwlStyle), uintptr(style))
	//C.window_add_border(w.handle)
}
