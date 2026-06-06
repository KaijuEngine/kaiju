/******************************************************************************/
/* mouse.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

import (
	"math"

	"kaijuengine.com/matrix"
)

const (
	MouseButtonLeft = iota
	MouseButtonMiddle
	MouseButtonRight
	MouseButtonX1
	MouseButtonX2
	MouseButtonLast
)

const (
	MouseRelease = iota
	MousePress
	MouseRepeat
	MouseInvalid            = -1
	MouseButtonStateInvalid = -1
)

type Mouse struct {
	X, Y             float32
	SX, SY           float32
	CX, CY           float32
	ScrollX, ScrollY float32
	buttonStates     [MouseButtonLast]int
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
		switch m.buttonStates[i] {
		case MouseRelease:
			m.buttonStates[i] = MouseButtonStateInvalid
		case MousePress:
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

func (m *Mouse) ForceHeld(index int) {
	m.buttonStates[index] = MouseRepeat
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
