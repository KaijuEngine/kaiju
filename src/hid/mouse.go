/******************************************************************************/
/* mouse.go                                                                   */
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

package hid

import (
	"kaiju/matrix"
	"kaiju/systems/events"
	"math"
)

const (
	MouseButtonLeft = iota
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonX1
	MouseButtonX2
	MouseButtonLast
	MouseInvalid = -1
	MouseRelease = iota
	MousePress
	MouseRepeat
	MouseButtonStateInvalid = -1
)

type DragData interface {
	DragUpdate()
}

type Mouse struct {
	X, Y             float32
	SX, SY           float32
	CX, CY           float32
	ScrollX, ScrollY float32
	buttonStates     [MouseButtonLast]int
	OnDragStop       events.Event
	dragData         DragData
	moved            bool
	buttonChanged    bool
	scrollPending    bool
}

func NewMouse() Mouse {
	m := Mouse{}
	for i := 0; i < MouseButtonLast; i++ {
		m.buttonStates[i] = MouseButtonStateInvalid
	}
	return m
}

func (m Mouse) Moved() bool {
	return m.moved
}

func (m Mouse) ButtonChanged() bool {
	return m.buttonChanged
}

func (m *Mouse) EndUpdate() {
	for i := 0; i < MouseButtonLast; i++ {
		if m.buttonStates[i] == MouseRelease {
			m.buttonStates[i] = MouseButtonStateInvalid
			m.SetDragData(nil)
		} else if m.buttonStates[i] == MousePress {
			m.buttonStates[i] = MouseRepeat
			m.buttonChanged = true
		}
	}
	m.ScrollX = 0.0
	m.ScrollY = 0.0
	m.moved = false
}

func (m *Mouse) SetPosition(x, y, windowWidth, windowHeight float32) {
	if m.X != x || m.Y != y {
		m.X = x
		m.Y = windowHeight - y
		m.SX = x
		m.SY = y
		m.CX = x - windowWidth/2.0
		m.CY = windowHeight/2.0 - y
		m.moved = true
		if m.dragData != nil {
			m.dragData.DragUpdate()
		}
	}
}

func (m *Mouse) SetDown(index int) {
	if m.buttonStates[index] == MouseInvalid {
		m.buttonStates[index] = MousePress
		m.buttonChanged = true
	}
}

func (m *Mouse) SetUp(index int) {
	if m.buttonStates[index] != MouseInvalid {
		m.buttonStates[index] = MouseRelease
		m.buttonChanged = true
	}
}

func (m Mouse) Pressed(index int) bool {
	if index > MouseButtonLast {
		return false
	}
	return m.buttonStates[index] == MousePress
}

func (m Mouse) Released(index int) bool {
	if index > MouseButtonLast {
		return false
	}
	return m.buttonStates[index] == MouseRelease
}

func (m Mouse) Held(index int) bool {
	if index > MouseButtonLast {
		return false
	}
	return m.buttonStates[index] == MouseRepeat
}

func (m Mouse) ButtonState(index int) int {
	if index > MouseButtonLast {
		return MouseButtonStateInvalid
	}
	return m.buttonStates[index]
}

func (m Mouse) Scrolled() bool {
	return matrix.Abs(m.ScrollY) >= math.SmallestNonzeroFloat32 ||
		matrix.Abs(m.ScrollX) >= math.SmallestNonzeroFloat32
}

func (m Mouse) Position() matrix.Vec2 {
	return matrix.Vec2{m.X, m.Y}
}

func (m Mouse) CenteredPosition() matrix.Vec2 {
	return matrix.Vec2{m.CX, m.CY}
}

func (m Mouse) ScreenPosition() matrix.Vec2 {
	return matrix.Vec2{m.SX, m.SY}
}

func (m Mouse) Scroll() matrix.Vec2 {
	return matrix.Vec2{m.ScrollX, m.ScrollY}
}

func (m *Mouse) SetScroll(x, y float32) {
	m.ScrollX = x
	m.ScrollY = y
	m.scrollPending = true
}

func (m *Mouse) Reset() {
	for i := 0; i < MouseButtonLast; i++ {
		if m.buttonStates[i] == MousePress || m.buttonStates[i] == MouseRepeat {
			m.buttonStates[i] = MouseRelease
		}
	}
}

func (m Mouse) DragData() any {
	return m.dragData
}

func (m *Mouse) SetDragData(data DragData) {
	if m.dragData != nil {
		m.OnDragStop.Execute()
	}
	m.dragData = data
}
