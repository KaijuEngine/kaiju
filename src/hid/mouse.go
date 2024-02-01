package hid

import (
	"kaiju/matrix"
	"math"
)

const (
	MouseButtonLeft         = 0
	MouseButtonMiddle       = 1
	MouseButtonRight        = 2
	MouseButtonX1           = 3
	MouseButtonX2           = 4
	MouseButtonLast         = 5
	MouseInvalid            = -1
	MouseRelease            = 0
	MousePress              = 1
	MouseRepeat             = 2
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
		if m.buttonStates[i] == MouseRelease {
			m.buttonStates[i] = MouseButtonStateInvalid
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
