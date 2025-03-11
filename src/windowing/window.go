/******************************************************************************/
/* window.go                                                                  */
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
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package windowing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"kaiju/assets"
	"kaiju/hid"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/profiler/tracing"
	"kaiju/rendering"
	"kaiju/systems/events"
	"slices"
	"unsafe"
)

type eventType = int

const (
	evtUnknown eventType = iota
	evtQuit
	evtMouseMove
	evtLeftMouseDown
	evtLeftMouseUp
	evtMiddleMouseDown
	evtMiddleMouseUp
	evtRightMouseDown
	evtRightMouseUp
	evtX1MouseDown
	evtX1MouseUp
	evtX2MouseDown
	evtX2MouseUp
	evtMouseWheelVertical
	evtMouseWheelHorizontal
	evtKeyDown
	evtKeyUp
	evtResize
	evtMove
	evtActivity
	evtControllerStates
)

var activeWindows []*Window

type Window struct {
	handle                   unsafe.Pointer
	instance                 unsafe.Pointer
	Mouse                    hid.Mouse
	Keyboard                 hid.Keyboard
	Touch                    hid.Touch
	Stylus                   hid.Stylus
	Controller               hid.Controller
	Cursor                   hid.Cursor
	Renderer                 rendering.Renderer
	evtSharedMem             *evtMem
	OnResize                 events.Event
	OnMove                   events.Event
	title                    string
	x, y                     int
	width, height            int
	left, top, right, bottom int // Full window including title and borders
	resetDragDataInFrames    int
	isClosed                 bool
	isCrashed                bool
	cursorChangeCount        int
	dragDropForceMouseUp     bool
}

type FileSearch struct {
	Title     string
	Extension string
}

func New(windowName string, width, height, x, y int, assets *assets.Database) (*Window, error) {
	w := &Window{
		Keyboard:     hid.NewKeyboard(),
		Mouse:        hid.NewMouse(),
		Touch:        hid.NewTouch(),
		Stylus:       hid.NewStylus(),
		Controller:   hid.NewController(),
		width:        width,
		height:       height,
		evtSharedMem: new(evtMem),
		x:            x,
		y:            y,
		title:        windowName,
	}
	activeWindows = slices.Insert(activeWindows, 0, w)
	w.Cursor = hid.NewCursor(&w.Mouse, &w.Touch, &w.Stylus)
	createWindow(windowName+"\x00\x00", w.width, w.height, x, y, w.evtSharedMem)
	if w.evtSharedMem.IsFatal() {
		return nil, errors.New(w.evtSharedMem.FatalMessage())
	}
	var hwndAddr, hInstance uint64
	reader := bytes.NewReader(w.evtSharedMem[evtSharedMemDataStart:])
	binary.Read(reader, binary.LittleEndian, &hwndAddr)
	w.handle = unsafe.Pointer(uintptr(hwndAddr))
	binary.Read(reader, binary.LittleEndian, &hInstance)
	w.instance = unsafe.Pointer(uintptr(hInstance))
	createWindowContext(w.handle, w.evtSharedMem)
	if w.evtSharedMem.IsFatal() {
		return nil, errors.New(w.evtSharedMem.FatalMessage())
	}
	w.showWindow(w.evtSharedMem)
	if w.evtSharedMem.IsFatal() {
		return nil, errors.New(w.evtSharedMem.FatalMessage())
	}
	var err error
	w.Renderer, err = selectRenderer(w, windowName, assets)
	w.x, w.y = w.position()
	return w, err
}

func FindWindowAtPoint(x, y int) (*Window, bool) {
	for i := range activeWindows {
		w := activeWindows[i]
		if x >= w.left && x <= w.right && y >= w.top && y <= w.bottom {
			return w, true
		}
	}
	return nil, false
}

func (w *Window) canChangeCursor() bool { return w.cursorChangeCount == 0 }

func (w *Window) ToScreenPosition(x, y int) (int, int) {
	leftBorder := (w.right - w.left - w.width) / 2
	topBorder := (w.bottom - w.top - w.height) - leftBorder // Borders are same?
	return x + (w.x + leftBorder), y + (w.y + topBorder)
}

func (w *Window) ToLocalPosition(x, y int) (int, int) {
	leftBorder := (w.right - w.left - w.width) / 2
	topBorder := (w.bottom - w.top - w.height) - leftBorder // Borders are same?
	return x - (w.x + leftBorder), y - (w.y + topBorder)
}

func (w *Window) PlatformWindow() unsafe.Pointer   { return w.cHandle() }
func (w *Window) PlatformInstance() unsafe.Pointer { return w.cInstance() }

func (w *Window) IsClosed() bool  { return w.isClosed }
func (w *Window) IsCrashed() bool { return w.isCrashed }
func (w *Window) X() int          { return w.x }
func (w *Window) Y() int          { return w.y }
func (w *Window) XY() (int, int)  { return w.x, w.y }
func (w *Window) Width() int      { return w.width }
func (w *Window) Height() int     { return w.height }

func (w *Window) Viewport() matrix.Vec4 {
	return matrix.Vec4{0, 0, float32(w.width), float32(w.height)}
}

func (w *Window) processEvent(evtType eventType) {
	w.processWindowEvent(evtType)
	w.processMouseEvent(evtType)
	w.processKeyboardEvent(evtType)
	w.processControllerEvent(evtType)
}

func (w *Window) processWindowEvent(evtType eventType) {
	switch evtType {
	case evtResize:
		we := w.evtSharedMem.toWindowResizeEvent()
		w.width = int(we.width)
		w.height = int(we.height)
		w.left = int(we.left)
		w.top = int(we.top)
		w.right = int(we.right)
		w.bottom = int(we.bottom)
		if w.Renderer != nil {
			w.Renderer.Resize(w.width, w.height)
		}
		w.OnResize.Execute()
	case evtMove:
		we := w.evtSharedMem.toWindowMoveEvent()
		// When a window is created at a specific position, windows will
		// sometimes report a move to position of 0,0 for some reason.
		if we.x != 0 || we.y != 0 {
			ww := w.right - w.left
			wh := w.bottom - w.top
			w.x = int(we.x)
			w.y = int(we.y)
			w.left = w.x
			w.top = w.y
			w.right = w.x + ww
			w.bottom = w.y + wh
		}
		w.OnMove.Execute()
	case evtActivity:
		ee := w.evtSharedMem.toEnumEvent()
		switch ee.value {
		case 0:
			w.becameInactive()
		case 1:
			w.becameActive()
		}
	}
}

func (w *Window) processMouseEvent(evtType eventType) {
	switch evtType {
	case evtMouseMove:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetPosition(float32(me.x), float32(me.y), float32(w.width), float32(w.height))
		UpdateDragData(w, int(me.x), int(me.y))
	case evtLeftMouseDown:
		w.Mouse.SetDown(hid.MouseButtonLeft)
	case evtLeftMouseUp:
		w.Mouse.SetUp(hid.MouseButtonLeft)
		w.resetDragDataInFrames = 2
		UpdateDragDrop(w, int(w.Mouse.SX), int(w.Mouse.SY))
	case evtMiddleMouseDown:
		w.Mouse.SetDown(hid.MouseButtonMiddle)
	case evtMiddleMouseUp:
		w.Mouse.SetUp(hid.MouseButtonMiddle)
	case evtRightMouseDown:
		w.Mouse.SetDown(hid.MouseButtonRight)
	case evtRightMouseUp:
		w.Mouse.SetUp(hid.MouseButtonRight)
	case evtX1MouseDown:
		me := w.evtSharedMem.toMouseEvent()
		if me.buttonId == 4 {
			w.Mouse.SetDown(hid.MouseButtonX2)
		} else {
			w.Mouse.SetDown(hid.MouseButtonX1)
		}
	case evtX1MouseUp:
		me := w.evtSharedMem.toMouseEvent()
		if me.buttonId == 4 {
			w.Mouse.SetUp(hid.MouseButtonX2)
		} else {
			w.Mouse.SetUp(hid.MouseButtonX1)
		}
	case evtX2MouseDown:
		w.Mouse.SetDown(hid.MouseButtonX2)
	case evtX2MouseUp:
		w.Mouse.SetUp(hid.MouseButtonX2)
	case evtMouseWheelVertical:
		s := w.Mouse.Scroll()
		me := w.evtSharedMem.toMouseEvent()
		delta := scaleScrollDelta(float32(me.delta))
		w.Mouse.SetScroll(s.X(), s.Y()+delta)
	case evtMouseWheelHorizontal:
		s := w.Mouse.Scroll()
		me := w.evtSharedMem.toMouseEvent()
		delta := scaleScrollDelta(float32(me.delta))
		w.Mouse.SetScroll(s.X()+delta, s.Y())
	}
}

func (w *Window) processKeyboardEvent(evtType eventType) {
	switch evtType {
	case evtKeyDown:
		ke := w.evtSharedMem.toKeyboardEvent()
		key := hid.ToKeyboardKey(int(ke.key))
		w.Keyboard.SetKeyDown(key)
	case evtKeyUp:
		ke := w.evtSharedMem.toKeyboardEvent()
		key := hid.ToKeyboardKey(int(ke.key))
		w.Keyboard.SetKeyUp(key)
	}
}

func (w *Window) processControllerEvent(evtType eventType) {
	switch evtType {
	case evtControllerStates:
		ce := w.evtSharedMem.toControllerEvent()
		for id := range ce.controllerStates {
			c := &ce.controllerStates[id]
			if c.isConnected == 0 {
				w.Controller.Disconnected(id)
			} else {
				w.Controller.Connected(id)
			}
			for i := 0; i < int(unsafe.Sizeof(c.buttons)*8); i++ {
				buttonId := c.buttons & (1 << i)
				if buttonId != 0 {
					w.Controller.SetButtonDown(id, i)
				} else {
					w.Controller.SetButtonUp(id, i)
				}
			}
		}
	}
}

func (w *Window) Poll() {
	defer tracing.NewRegion("Window::Poll").End()
	w.poll()
	w.isClosed = w.isClosed || w.evtSharedMem.IsQuit()
	w.isCrashed = w.isCrashed || w.evtSharedMem.IsFatal()
	w.Cursor.Poll()
	if w.dragDropForceMouseUp {
		w.Mouse.SetUp(hid.MouseButtonLeft)
		w.dragDropForceMouseUp = false
	}
}

func (w *Window) EndUpdate() {
	defer tracing.NewRegion("Window::EndUpdate").End()
	w.Keyboard.EndUpdate()
	w.Mouse.EndUpdate()
	w.Touch.EndUpdate()
	w.Stylus.EndUpdate()
	w.Controller.EndUpdate()
	if w.resetDragDataInFrames > 0 {
		// We wait a number of frames to allow for cross-window communication
		w.resetDragDataInFrames--
		if w.resetDragDataInFrames == 0 {
			SetDragData(nil)
		}
	}
}

func (w *Window) SwapBuffers() {
	defer tracing.NewRegion("Window::SwapBuffers").End()
	if w.Renderer.SwapFrame(int32(w.Width()), int32(w.Height())) {
		swapBuffers(w.handle)
	}
}

func (w *Window) SizeMM() (int, int, error) {
	return w.sizeMM()
}

func (w *Window) IsPhoneSize() bool {
	wmm, hmm, _ := w.SizeMM()
	return wmm < 178 || hmm < 170
}

func (w *Window) IsPCSize() bool {
	wmm, hmm, _ := w.SizeMM()
	return wmm > 254 || hmm > 254
}

func (w *Window) IsTabletSize() bool {
	return !w.IsPhoneSize() && !w.IsPCSize()
}

func DPI2PX(pixels, mm, targetMM int) int {
	return targetMM * (pixels / mm)
}

func (w *Window) CursorStandard() {
	w.cursorChangeCount--
	if w.cursorChangeCount == 0 {
		w.cursorStandard()
	}
}

func (w *Window) CursorIbeam() {
	if w.canChangeCursor() {
		w.cursorIbeam()
	}
	w.cursorChangeCount++
}

func (w *Window) CursorSizeAll() {
	if w.canChangeCursor() {
		w.cursorSizeAll()
	}
	w.cursorChangeCount++
}

func (w *Window) CursorSizeNS() {
	if w.canChangeCursor() {
		w.cursorSizeNS()
	}
	w.cursorChangeCount++
}

func (w *Window) CursorSizeWE() {
	if w.canChangeCursor() {
		w.cursorSizeWE()
	}
	w.cursorChangeCount++
}

func (w *Window) CopyToClipboard(text string) { w.copyToClipboard(text) }
func (w *Window) ClipboardContents() string   { return w.clipboardContents() }

func (w *Window) removeFromActiveWindows() {
	for i := range activeWindows {
		if activeWindows[i] == w {
			activeWindows = slices.Delete(activeWindows, i, i+1)
			break
		}
	}
}

func (w *Window) Destroy() {
	w.isClosed = true
	w.Renderer.Destroy()
	w.destroy()
	w.removeFromActiveWindows()
}

func (w *Window) Focus() {
	w.focus()
	w.cursorStandard()
}

func (w *Window) Position() (x int, y int) {
	x, y = w.position()
	w.x = x
	w.y = y
	return x, y
}

func (w *Window) SetPosition(x, y int) {
	w.setPosition(x, y)
	w.x = x
	w.y = y
}

func (w *Window) SetSize(width, height int) {
	w.setSize(width, height)
	w.width = width
	w.height = height
}

func (w *Window) RemoveBorder() { w.removeBorder() }
func (w *Window) AddBorder()    { w.addBorder() }

func (w *Window) Center() (x int, y int) {
	x, y = w.Position()
	return x + w.Width()/2, y + w.Height()/2
}

func (w *Window) becameInactive() {
	w.Keyboard.Reset()
	w.Mouse.Reset()
	w.Touch.Reset()
	w.Stylus.Reset()
	w.Controller.Reset()
}

func (w *Window) becameActive() {
	w.cursorStandard()
	idx := -1
	for i := range activeWindows {
		if activeWindows[i] == w {
			idx = i
			break
		}
	}
	klib.SliceMove(activeWindows, idx, 0)
}
