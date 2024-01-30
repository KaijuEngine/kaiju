package windowing

import (
	"unsafe"
)

const (
	sharedMemAvailable = iota
	sharedMemWriting
	sharedMemWritten
	sharedMemAwaitingContext = 0xFC
	sharedMemAwaitingStart   = 0xFD
	sharedMemFatal           = 0xFE
	sharedMemQuit            = 0xFF
)

const (
	nativeMouseButtonLeft   = 0
	nativeMouseButtonMiddle = 1
	nativeMouseButtonRight  = 2
	nativeMouseButtonX1     = 3
	nativeMouseButtonX2     = 4
)

const (
	evtSharedMemSize      = 256
	evtSharedMemDataStart = 4
)

type evtMem [evtSharedMemSize]byte

type baseEvent struct {
	eventType uint32
}

type mouseEvent struct {
	baseEvent
	mouseButtonId int32
	mouseX        int32
	mouseY        int32
}

type windowEvent struct {
	baseEvent
	width  int32
	height int32
}

type keyboardEvent struct {
	baseEvent
	key int32
}

type controllerState struct {
	baseEvent
	buttons      uint16
	thumbLX      int16
	thumbLY      int16
	thumbRX      int16
	thumbRY      int16
	leftTrigger  uint8
	rightTrigger uint8
	isConnected  uint8
}

type controllerEvent struct {
	// TODO:  This 4 will need to be pulled from C
	controllerStates [4]controllerState
}

func (e *evtMem) AsPointer() unsafe.Pointer     { return unsafe.Pointer(&e[0]) }
func (e *evtMem) AsDataPointer() unsafe.Pointer { return unsafe.Pointer(&e[evtSharedMemDataStart]) }
func (e evtMem) IsFatal() bool                  { return e[0] == sharedMemFatal }
func (e evtMem) FatalMessage() string           { return string([]byte(e[evtSharedMemDataStart:])) }
func (e evtMem) IsReady() bool                  { return e[0] >= sharedMemWritten }
func (e evtMem) IsStart() bool                  { return e[0] == sharedMemAwaitingStart }
func (e evtMem) IsContext() bool                { return e[0] == sharedMemAwaitingContext }
func (e evtMem) IsWritten() bool                { return e[0] == sharedMemWritten }
func (e evtMem) IsQuit() bool                   { return e[0] == sharedMemQuit }
func (e *evtMem) MakeAvailable()                { e[0] = sharedMemAvailable }
func (e evtMem) HasEvent() bool                 { return e.EventType() != 0 }
func (e evtMem) EventType() uint32 {
	return *(*uint32)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}
func (e *evtMem) SetFatal(message string) {
	e[0] = sharedMemFatal
	msg := []byte(message)
	for i := 0; i < len(msg) && i < len(e)-evtSharedMemDataStart; i++ {
		e[i+evtSharedMemDataStart] = msg[i]
	}
}
func (e *evtMem) AwaitReady() {
	for !e.IsReady() {
	}
}

func (e evtMem) toWindowEvent() *windowEvent {
	return (*windowEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e evtMem) toMouseEvent() *mouseEvent {
	return (*mouseEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e evtMem) toKeyboardEvent() *keyboardEvent {
	return (*keyboardEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e evtMem) toControllerEvent() *controllerEvent {
	return (*controllerEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}
