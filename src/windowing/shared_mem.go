/*****************************************************************************/
/* shared_mem.go                                                             */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package windowing

import (
	"unsafe"
)

const (
	sharedMemResize          = 0xFB
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
	buttonId int32
	x        int32
	y        int32
	delta    int32
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
func (e *evtMem) IsFatal() bool                 { return e[0] == sharedMemFatal }
func (e *evtMem) FatalMessage() string          { return string([]byte(e[evtSharedMemDataStart:])) }
func (e *evtMem) IsQuit() bool                  { return e[0] == sharedMemQuit }
func (e *evtMem) IsResize() bool                { return e[0] == sharedMemResize }
func (e *evtMem) ResetHeader()                  { e[0] = 0 }

func (e *evtMem) SetFatal(message string) {
	e[0] = sharedMemFatal
	msg := []byte(message)
	for i := 0; i < len(msg) && i < len(e)-evtSharedMemDataStart; i++ {
		e[i+evtSharedMemDataStart] = msg[i]
	}
}

func (e *evtMem) toWindowEvent() *windowEvent {
	return (*windowEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e *evtMem) toMouseEvent() *mouseEvent {
	return (*mouseEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e *evtMem) toKeyboardEvent() *keyboardEvent {
	return (*keyboardEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}

func (e *evtMem) toControllerEvent() *controllerEvent {
	return (*controllerEvent)(unsafe.Pointer(&e[unsafe.Sizeof(uint32(0))]))
}
