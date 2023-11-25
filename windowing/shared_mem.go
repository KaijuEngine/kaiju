package windowing

import (
	"unsafe"
)

const (
	sharedMemAvailable = iota
	sharedMemWriting
	sharedMemWritten
	sharedMemFatal = 0xFE
	sharedMemQuit  = 0xFF
)

const (
	nativeMouseButtonLeft   = 0
	nativeMouseButtonMiddle = 1
	nativeMouseButtonRight  = 2
	nativeMouseButtonX1     = 3
	nativeMouseButtonX2     = 4
)

const evtSharedMemSize = 256

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

type keyboardEvent struct {
	baseEvent
	key int32
}

func (e *evtMem) AsPointer() unsafe.Pointer { return unsafe.Pointer(&e[0]) }
func (e evtMem) IsFatal() bool              { return e[0] == sharedMemFatal }
func (e evtMem) IsReady() bool              { return e[0] >= sharedMemWritten }
func (e evtMem) IsWritten() bool            { return e[0] == sharedMemWritten }
func (e evtMem) IsQuit() bool               { return e[0] == sharedMemQuit }
func (e *evtMem) MakeAvailable()            { e[0] = sharedMemAvailable }
func (e evtMem) HasEvent() bool             { return e.EventType() != 0 }
func (e evtMem) EventType() uint32 {
	return *(*uint32)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e evtMem) toMouseEvent() *mouseEvent {
	return (*mouseEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e evtMem) toKeyboardEvent() *keyboardEvent {
	return (*keyboardEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}
