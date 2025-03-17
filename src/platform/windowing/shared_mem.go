/******************************************************************************/
/* shared_mem.go                                                              */
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

import "unsafe"

type WindowEventType = uint8
type WindowEventActivityType = uint32
type WindowEventButtonType = uint32
type WindowEventControllerConnectionType = uint32

const (
	sharedMemWindowActivity  = 0xF9
	sharedMemWindowMove      = 0xFA
	sharedMemWindowResize    = 0xFB
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
	windowEventTypeSetHandle       = WindowEventType(1)
	windowEventTypeActivity        = WindowEventType(2)
	windowEventTypeMove            = WindowEventType(3)
	windowEventTypeResize          = WindowEventType(4)
	windowEventTypeMouseMove       = WindowEventType(5)
	windowEventTypeMouseScroll     = WindowEventType(6)
	windowEventTypeMouseButton     = WindowEventType(7)
	windowEventTypeKeyboardButton  = WindowEventType(8)
	windowEventTypeControllerState = WindowEventType(9)
	windowEventTypeFatal           = WindowEventType(10)
)

const (
	windowEventActivityTypeMinimize = WindowEventActivityType(1)
	windowEventActivityTypeMaximize = WindowEventActivityType(2)
	windowEventActivityTypeClose    = WindowEventActivityType(3)
	windowEventActivityTypeFocus    = WindowEventActivityType(4)
	windowEventActivityTypeBlur     = WindowEventActivityType(5)
)

const (
	windowEventButtonTypeDown = WindowEventButtonType(1)
	windowEventButtonTypeUp   = WindowEventButtonType(2)
)

const (
	windowEventControllerConnectionTypeDisconnected = WindowEventControllerConnectionType(1)
	windowEventControllerConnectionTypeConnected    = WindowEventControllerConnectionType(2)
)

type SetHandleEvent struct {
	hwnd     unsafe.Pointer
	instance unsafe.Pointer
}

type WindowActivityEvent struct {
	activityType WindowEventActivityType
	_            [4]byte
}

type WindowMoveEvent struct {
	x int32
	y int32
}

type WindowResizeEvent struct {
	width  int32
	height int32
	left   int32
	top    int32
	right  int32
	bottom int32
}

type MouseMoveWindowEvent struct {
	x int32
	y int32
}

type MouseScrollWindowEvent struct {
	deltaX int32
	deltaY int32
	x      int32
	y      int32
}

type MouseButtonWindowEvent struct {
	buttonId int32
	action   WindowEventButtonType
	x        int32
	y        int32
}

type KeyboardButtonWindowEvent struct {
	buttonId int32
	action   WindowEventButtonType
}

type ControllerStateWindowEvent struct {
	controllerId   uint8
	leftTrigger    uint8
	rightTrigger   uint8
	_              byte
	connectionType WindowEventControllerConnectionType
	buttons        uint16
	thumbLX        int16
	thumbLY        int16
	thumbRX        int16
	thumbRY        int16
	_              [6]byte
}

const evtUnionSize = max(
	unsafe.Sizeof(SetHandleEvent{}),
	unsafe.Sizeof(WindowActivityEvent{}),
	unsafe.Sizeof(WindowMoveEvent{}),
	unsafe.Sizeof(WindowResizeEvent{}),
	unsafe.Sizeof(MouseMoveWindowEvent{}),
	unsafe.Sizeof(MouseScrollWindowEvent{}),
	unsafe.Sizeof(MouseButtonWindowEvent{}),
	unsafe.Sizeof(KeyboardButtonWindowEvent{}),
	unsafe.Sizeof(ControllerStateWindowEvent{}),
)

func readType(head unsafe.Pointer) (WindowEventType, unsafe.Pointer) {
	eType := WindowEventType(*(*uint64)(head))
	return eType, unsafe.Pointer(uintptr(head) + unsafe.Sizeof(uint64(0)))
}

func asSetHandleEvent(data unsafe.Pointer) *SetHandleEvent {
	return (*SetHandleEvent)(data)
}

func asWindowActivityEvent(data unsafe.Pointer) *WindowActivityEvent {
	return (*WindowActivityEvent)(data)
}

func asWindowMoveEvent(data unsafe.Pointer) *WindowMoveEvent {
	return (*WindowMoveEvent)(data)
}

func asWindowResizeEvent(data unsafe.Pointer) *WindowResizeEvent {
	return (*WindowResizeEvent)(data)
}

func asMouseMoveWindowEvent(data unsafe.Pointer) *MouseMoveWindowEvent {
	return (*MouseMoveWindowEvent)(data)
}

func asMouseScrollWindowEvent(data unsafe.Pointer) *MouseScrollWindowEvent {
	return (*MouseScrollWindowEvent)(data)
}

func asMouseButtonWindowEvent(data unsafe.Pointer) *MouseButtonWindowEvent {
	return (*MouseButtonWindowEvent)(data)
}

func asKeyboardButtonWindowEvent(data unsafe.Pointer) *KeyboardButtonWindowEvent {
	return (*KeyboardButtonWindowEvent)(data)
}

func asControllerStateWindowEvent(data unsafe.Pointer) *ControllerStateWindowEvent {
	return (*ControllerStateWindowEvent)(data)
}
