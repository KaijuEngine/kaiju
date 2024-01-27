package windowing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"kaiju/hid"
	"kaiju/rendering"
	"unsafe"
)

type eventType int

const (
	evtUnknown eventType = iota
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
)

type Window struct {
	handle        unsafe.Pointer
	instance      unsafe.Pointer
	Mouse         hid.Mouse
	Keyboard      hid.Keyboard
	Touch         hid.Touch
	Stylus        hid.Stylus
	Cursor        hid.Cursor
	Renderer      rendering.Renderer
	evtSharedMem  *evtMem
	width, height int
	isClosed      bool
	isCrashed     bool
}

func New(windowName string) (*Window, error) {
	w := &Window{
		Keyboard:     hid.NewKeyboard(),
		Mouse:        hid.NewMouse(),
		Touch:        hid.NewTouch(),
		Stylus:       hid.NewStylus(),
		width:        944,
		height:       500,
		evtSharedMem: new(evtMem),
	}
	w.Cursor = hid.NewCursor(&w.Mouse, &w.Touch, &w.Stylus)
	// TODO:  Pass in width and height
	createWindow(windowName, w.width, w.height, w.evtSharedMem)
	w.evtSharedMem.AwaitReady()
	if !w.evtSharedMem.IsFatal() && !w.evtSharedMem.IsContext() {
		return nil, errors.New("Context create expected but wasn't requested")
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
	w.evtSharedMem.MakeAvailable()
	w.evtSharedMem.AwaitReady()
	if w.evtSharedMem.IsFatal() {
		return nil, errors.New(w.evtSharedMem.FatalMessage())
	} else if !w.evtSharedMem.IsStart() {
		return nil, errors.New("Start expected but wasn't requested")
	}
	var err error
	w.Renderer, err = selectRenderer(w, windowName)
	return w, err
}

func (w Window) PlatformWindow() unsafe.Pointer   { return w.handle }
func (w Window) PlatformInstance() unsafe.Pointer { return w.instance }

func (w Window) IsClosed() bool  { return w.isClosed }
func (w Window) IsCrashed() bool { return w.isCrashed }
func (w Window) Width() int      { return w.width }
func (w Window) Height() int     { return w.height }

func (w *Window) processEvent() {
	evtType := w.evtSharedMem.toEventType()
	w.processMouseEvent(evtType)
	w.processKeyboardEvent(evtType)
}

func (w *Window) processMouseEvent(evtType eventType) {
	switch evtType {
	case evtMouseMove:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetPosition(float32(me.mouseX), float32(me.mouseY), float32(w.width), float32(w.height))
	case evtLeftMouseDown:
		w.Mouse.SetDown(hid.MouseButtonLeft)
	case evtLeftMouseUp:
		w.Mouse.SetUp(hid.MouseButtonLeft)
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
		if me.mouseButtonId == 4 {
			println("X2 down")
			w.Mouse.SetDown(hid.MouseButtonX2)
		} else {
			println("X1 down")
			w.Mouse.SetDown(hid.MouseButtonX1)
		}
	case evtX1MouseUp:
		me := w.evtSharedMem.toMouseEvent()
		if me.mouseButtonId == 4 {
			println("X2 up")
			w.Mouse.SetUp(hid.MouseButtonX2)
		} else {
			println("X1 up")
			w.Mouse.SetUp(hid.MouseButtonX1)
		}
	case evtX2MouseDown:
		w.Mouse.SetDown(hid.MouseButtonX2)
	case evtX2MouseUp:
		w.Mouse.SetUp(hid.MouseButtonX2)
	case evtMouseWheelVertical:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetScroll(0.0, float32(me.mouseY))
	case evtMouseWheelHorizontal:
		me := w.evtSharedMem.toMouseEvent()
		w.Mouse.SetScroll(float32(me.mouseX), 0.0)
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

func (w *Window) Poll() {
	w.evtSharedMem.MakeAvailable()
	for !w.evtSharedMem.IsQuit() && !w.evtSharedMem.IsFatal() {
		for !w.evtSharedMem.IsReady() {
		}
		if w.evtSharedMem.IsWritten() {
			if w.evtSharedMem.HasEvent() {
				w.processEvent()
				w.evtSharedMem.MakeAvailable()
			} else {
				break
			}
		}
	}
	w.isClosed = w.isClosed || w.evtSharedMem.IsQuit()
	w.isCrashed = w.isCrashed || w.evtSharedMem.IsFatal()
	w.Cursor.Poll()
}

func (w *Window) EndUpdate() {
	w.Keyboard.EndUpdate()
	w.Mouse.EndUpdate()
	w.Touch.EndUpdate()
	w.Stylus.EndUpdate()
}

func (w *Window) SwapBuffers() {
	w.Renderer.SwapFrame(int32(w.Width()), int32(w.Height()))
	swapBuffers(w.handle)
}

func (w Window) GetDPI() (int, int, error) {
	// TODO:  Actually get the DPI
	return 96, 96, nil
}

func (w Window) IsPhoneSize() bool {
	wmm, hmm, _ := w.GetDPI()
	return wmm < 178 || hmm < 170
}

func (w Window) IsPCSize() bool {
	wmm, hmm, _ := w.GetDPI()
	return wmm > 254 || hmm > 254
}

func (w Window) IsTabletSize() bool {
	return !w.IsPhoneSize() && !w.IsPCSize()
}

func DPI2PX(pixels, mm, targetMM int) int {
	return targetMM * (pixels / mm)
}
