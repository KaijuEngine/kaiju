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
	"errors"
	"kaiju/engine/assets"
	"kaiju/engine/systems/events"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/filesystem"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/rendering"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	activeWindows []*Window
	windowLookup  sync.Map
	nextLookupId  atomic.Uint64
)

type Window struct {
	lookupId                 uint64
	handle                   unsafe.Pointer
	instance                 unsafe.Pointer
	Mouse                    hid.Mouse
	Keyboard                 hid.Keyboard
	Touch                    hid.Touch
	Stylus                   hid.Stylus
	Controller               hid.Controller
	Cursor                   hid.Cursor
	Renderer                 rendering.Renderer
	OnResize                 events.Event
	OnMove                   events.Event
	title                    string
	x, y                     int
	width, height            int
	left, top, right, bottom int // Full window including title and borders
	resetDragDataInFrames    int
	cursorChangeCount        int
	windowSync               chan struct{}
	syncRequest              bool
	isClosed                 bool
	isCrashed                bool
	fatalFromNativeAPI       bool
	resizedFromNativeAPI     bool
	isFullScreen             bool
}

type FileSearch struct {
	Title     string
	Extension string
}

type WindowCleanup struct {
	handle   unsafe.Pointer
	renderer rendering.Renderer
}

func New(windowName string, width, height, x, y int, assets *assets.Database) (*Window, error) {
	defer tracing.NewRegion("windowing.New").End()
	w := &Window{
		Keyboard:   hid.NewKeyboard(),
		Mouse:      hid.NewMouse(),
		Touch:      hid.NewTouch(),
		Stylus:     hid.NewStylus(),
		Controller: hid.NewController(),
		width:      width,
		height:     height,
		x:          x,
		y:          y,
		left:       x,
		top:        y,
		right:      x + width,
		bottom:     y + height,
		title:      windowName,
		windowSync: make(chan struct{}),
	}
	activeWindows = slices.Insert(activeWindows, 0, w)
	w.Cursor = hid.NewCursor(&w.Mouse, &w.Touch, &w.Stylus)
	w.createWindow(windowName+"\x00\x00", x, y)
	if w.fatalFromNativeAPI {
		return nil, errors.New("failed to create the window " + windowName)
	}
	createWindowContext(w.handle)
	if w.fatalFromNativeAPI {
		return nil, errors.New("failed to create the window context for " + windowName)
	}
	w.showWindow()
	if w.fatalFromNativeAPI {
		return nil, errors.New("failed to present the window " + windowName)
	}
	var err error
	w.Renderer, err = selectRenderer(w, windowName, assets)
	w.x, w.y = w.position()
	runtime.AddCleanup(w, func(state WindowCleanup) {
		state.renderer.Destroy()
		destroyWindow(state.handle)
	}, WindowCleanup{w.handle, w.Renderer})
	return w, err
}

func FindWindowAtPoint(x, y int) (*Window, bool) {
	defer tracing.NewRegion("windowing.FindWindowAtPoint").End()
	for i := range activeWindows {
		w := activeWindows[i]
		if x >= w.left && x <= w.right && y >= w.top && y <= w.bottom {
			return w, true
		}
	}
	return nil, false
}

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

func (w *Window) Poll() {
	defer tracing.NewRegion("Window.Poll").End()
	if w.syncRequest {
		w.windowSync <- struct{}{}
		<-w.windowSync
		w.syncRequest = false
	}
	w.poll()
	if w.resizedFromNativeAPI {
		w.resizedFromNativeAPI = false
		if w.Renderer != nil {
			w.Renderer.Resize(w.width, w.height)
		}
		w.OnResize.Execute()
	}
	w.isCrashed = w.isCrashed || w.fatalFromNativeAPI
	w.Cursor.Poll()
}

func (w *Window) EndUpdate() {
	defer tracing.NewRegion("Window.EndUpdate").End()
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
	defer tracing.NewRegion("Window.SwapBuffers").End()
	if w.Renderer.SwapFrame(int32(w.Width()), int32(w.Height())) {
		swapBuffers(w.handle)
	}
}

func (w *Window) DotsPerMillimeter() float64 {
	return w.dotsPerMillimeter()
}

func (w *Window) SizeMM() (int, int, error) {
	return w.sizeMM()
}

func (w *Window) ScreenSizeMM() (int, int, error) {
	return w.screenSizeMM()
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

func DPI2PXF(pixels, mm, targetMM float64) float64 {
	return targetMM * (pixels / mm)
}

func (w *Window) CursorStandard() {
	w.cursorChangeCount = max(0, w.cursorChangeCount-1)
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

func (w *Window) Destroy() {
	defer tracing.NewRegion("Window.Destroy").End()
	w.isClosed = true
	w.removeFromActiveWindows()
	w.Renderer.Teardown()
}

func (w *Window) Focus() {
	defer tracing.NewRegion("Window.Focus").End()
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

func (w *Window) RemoveBorder()      { w.removeBorder() }
func (w *Window) AddBorder()         { w.addBorder() }
func (w *Window) ShowCursor()        { w.showCursor() }
func (w *Window) HideCursor()        { w.hideCursor() }
func (w *Window) IsFullScreen() bool { return w.isFullScreen }

func (w *Window) SetFullscreen() {
	if w.isFullScreen {
		return
	}
	w.setFullscreen()
	w.isFullScreen = true
}

func (w *Window) SetWindowed(width, height int) {
	w.setWindowed(width, height)
	w.isFullScreen = false
}

func (w *Window) Center() (x int, y int) {
	x, y = w.Position()
	return x + w.Width()/2, y + w.Height()/2
}

func (w *Window) OpenFileDialog(startPath string, extensions []filesystem.DialogExtension, ok func(path string), cancel func()) error {
	w.disableRawMouse()
	return filesystem.OpenFileDialogWindow(startPath, extensions, func(path string) {
		w.enableRawMouse()
		ok(path)
	}, func() {
		w.enableRawMouse()
		if cancel != nil {
			cancel()
		}
	}, w.handle)
}

func (w *Window) SaveFileDialog(startPath string, fileName string, extensions []filesystem.DialogExtension, ok func(path string), cancel func()) error {
	w.disableRawMouse()
	return filesystem.OpenSaveFileDialogWindow(startPath, fileName, extensions, func(path string) {
		w.enableRawMouse()
		ok(path)
	}, func() {
		w.enableRawMouse()
		if cancel != nil {
			cancel()
		}
	}, w.handle)
}

func (w *Window) EnableRawMouseInput()  { w.enableRawMouse() }
func (w *Window) DisableRawMouseInput() { w.disableRawMouse() }

func (w *Window) requestSync() {
	w.syncRequest = true
}

func (w *Window) canChangeCursor() bool { return w.cursorChangeCount == 0 }

func (w *Window) processWindowResizeEvent(evt *WindowResizeEvent) {
	w.width = int(evt.width)
	w.height = int(evt.height)
	w.left = int(evt.left)
	w.top = int(evt.top)
	w.right = int(evt.right)
	w.bottom = int(evt.bottom)
}

func (w *Window) processWindowMoveEvent(evt *WindowMoveEvent) {
	ww := w.right - w.left
	wh := w.bottom - w.top
	w.x = int(evt.x)
	w.y = int(evt.y)
	w.left = w.x
	w.top = w.y
	w.right = w.x + ww
	w.bottom = w.y + wh
	w.OnMove.Execute()
}

func (w *Window) processWindowActivityEvent(evt *WindowActivityEvent) {
	defer tracing.NewRegion("Window.processWindowActivityEvent").End()
	switch evt.activityType {
	case windowEventActivityTypeMinimize:
		// TODO:  Not implemented yet
	case windowEventActivityTypeMaximize:
		// TODO:  Not implemented yet
	case windowEventActivityTypeClose:
		w.isClosed = true
	case windowEventActivityTypeFocus:
		w.becameActive()
	case windowEventActivityTypeBlur:
		w.becameInactive()
	}
}

func (w *Window) processMouseMoveEvent(evt *MouseMoveWindowEvent) {
	defer tracing.NewRegion("Window.processMouseMoveEvent").End()
	w.Mouse.SetPosition(float32(evt.x), float32(evt.y), float32(w.width), float32(w.height))
	UpdateDragData(w, int(evt.x), int(evt.y))
}

func (w *Window) processMouseButtonEvent(evt *MouseButtonWindowEvent) {
	defer tracing.NewRegion("Window.processMouseButtonEvent").End()
	var targetButton int
	switch evt.buttonId {
	case nativeMouseButtonLeft:
		targetButton = hid.MouseButtonLeft
	case nativeMouseButtonMiddle:
		targetButton = hid.MouseButtonMiddle
	case nativeMouseButtonRight:
		targetButton = hid.MouseButtonRight
	case nativeMouseButtonX1:
		targetButton = hid.MouseButtonX1
	case nativeMouseButtonX2:
		targetButton = hid.MouseButtonX2
	}
	switch evt.action {
	case windowEventButtonTypeDown:
		w.Mouse.SetDown(targetButton)
	case windowEventButtonTypeUp:
		w.Mouse.SetUp(targetButton)
		if targetButton == hid.MouseButtonLeft {
			w.resetDragDataInFrames = 2
			UpdateDragDrop(w, int(w.Mouse.SX), int(w.Mouse.SY))
		}
	}
}

func (w *Window) processMouseScrollEvent(evt *MouseScrollWindowEvent) {
	defer tracing.NewRegion("Window.processMouseScrollEvent").End()
	s := w.Mouse.Scroll()
	deltaX := scaleScrollDelta(float32(evt.deltaX))
	w.Mouse.SetScroll(s.X(), s.Y()+deltaX)
	deltaY := scaleScrollDelta(float32(evt.deltaY))
	w.Mouse.SetScroll(s.X(), s.Y()+deltaY)
}

func (w *Window) processKeyboardButtonEvent(evt *KeyboardButtonWindowEvent) {
	defer tracing.NewRegion("Window.processKeyboardButtonEvent").End()
	switch evt.action {
	case windowEventButtonTypeDown:
		key := hid.ToKeyboardKey(int(evt.buttonId))
		w.Keyboard.SetKeyDown(key)
	case windowEventButtonTypeUp:
		key := hid.ToKeyboardKey(int(evt.buttonId))
		w.Keyboard.SetKeyUp(key)
	}
}

func (w *Window) processControllerStateEvent(evt *ControllerStateWindowEvent) {
	defer tracing.NewRegion("Window.processControllerStateEvent").End()
	if evt.connectionType == windowEventControllerConnectionTypeDisconnected {
		w.Controller.Disconnected(int(evt.controllerId))
	} else {
		w.Controller.Connected(int(evt.controllerId))
	}
	for i := 0; i < int(unsafe.Sizeof(evt.buttons)*8); i++ {
		buttonId := evt.buttons & (1 << i)
		if buttonId != 0 {
			w.Controller.SetButtonDown(int(evt.controllerId), i)
		} else {
			w.Controller.SetButtonUp(int(evt.controllerId), i)
		}
	}
}

func (w *Window) removeFromActiveWindows() {
	defer tracing.NewRegion("Window.removeFromActiveWindows").End()
	for i := range activeWindows {
		if activeWindows[i] == w {
			activeWindows = slices.Delete(activeWindows, i, i+1)
			break
		}
	}
	windowLookup.Delete(w.lookupId)
}

func (w *Window) becameInactive() {
	defer tracing.NewRegion("Window.becameInactive").End()
	w.Keyboard.Reset()
	w.Mouse.Reset()
	w.Touch.Reset()
	w.Stylus.Reset()
	w.Controller.Reset()
}

func (w *Window) becameActive() {
	defer tracing.NewRegion("Window.becameActive").End()
	w.cursorStandard()
	idx := -1
	for i := range activeWindows {
		if activeWindows[i] == w {
			idx = i
			break
		}
	}
	if idx >= 0 {
	klib.SliceMove(activeWindows, idx, 0)
	}
}

func goProcessEventsCommon(goWindow uint64, events unsafe.Pointer, eventCount uint32) {
	defer tracing.NewRegion("windowing.goProcessEventsCommon").End()
	gw, ok := windowLookup.Load(goWindow)
	if !ok || gw == nil {
		return
	}
	win := gw.(*Window)
	for range eventCount {
		eType, body := readType(events)
		switch eType {
		case windowEventTypeSetHandle:
			evt := asSetHandleEvent(body)
			win.handle = evt.hwnd
			win.instance = evt.instance
		case windowEventTypeActivity:
			win.processWindowActivityEvent(asWindowActivityEvent(body))
		case windowEventTypeMove:
			win.processWindowMoveEvent(asWindowMoveEvent(body))
		case windowEventTypeResize:
			win.processWindowResizeEvent(asWindowResizeEvent(body))
			win.resizedFromNativeAPI = true
		case windowEventTypeMouseMove:
			win.processMouseMoveEvent(asMouseMoveWindowEvent(body))
		case windowEventTypeMouseScroll:
			win.processMouseScrollEvent(asMouseScrollWindowEvent(body))
		case windowEventTypeMouseButton:
			win.processMouseButtonEvent(asMouseButtonWindowEvent(body))
		case windowEventTypeKeyboardButton:
			win.processKeyboardButtonEvent(asKeyboardButtonWindowEvent(body))
		case windowEventTypeControllerState:
			win.processControllerStateEvent(asControllerStateWindowEvent(body))
		case windowEventTypeFatal:
			events = body
			win.fatalFromNativeAPI = true
		}
		events = unsafe.Pointer(uintptr(body) + evtUnionSize)
	}
}
