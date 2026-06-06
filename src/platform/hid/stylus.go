/******************************************************************************/
/* stylus.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

type StylusActionState = int

const (
	StylusActionNone StylusActionState = iota
	StylusActionHoverEnter
	StylusActionHoverMove
	StylusActionHoverExit
	StylusActionDown
	StylusActionMove
	StylusActionUp
	StylusActionHeld
	StylusActionHover
)

type Stylus struct {
	X           float32
	Y           float32
	SX          float32
	SY          float32
	Pressure    float32
	Distance    float32
	actionState StylusActionState
}

func NewStylus() Stylus {
	return Stylus{
		actionState: StylusActionNone,
	}
}

func (s *Stylus) Pressed() bool {
	return s.actionState == StylusActionDown
}

func (s *Stylus) Moved() bool {
	return s.actionState == StylusActionMove || s.actionState == StylusActionHoverMove
}

func (s *Stylus) Held() bool {
	return s.actionState == StylusActionHeld || s.actionState == StylusActionMove
}

func (s *Stylus) Released() bool {
	return s.actionState == StylusActionUp
}

func (s *Stylus) ActionState() int {
	return s.actionState
}

func (s *Stylus) SetActionState(state StylusActionState) {
	s.actionState = state
}

func (s *Stylus) Set(x, y, pressure, distance, windowHeight float32) {
	s.X = x
	s.Y = windowHeight - y
	s.SX = x
	s.SY = y
	s.Pressure = pressure
	s.Distance = distance
}

func (s *Stylus) EndUpdate() {
	switch s.actionState {
	case StylusActionDown:
		fallthrough
	case StylusActionMove:
		s.actionState = StylusActionHeld
	case StylusActionUp:
		s.actionState = StylusActionNone
	case StylusActionHoverEnter:
		fallthrough
	case StylusActionHoverMove:
		s.actionState = StylusActionHover
	case StylusActionHoverExit:
		s.actionState = StylusActionNone
	}
}

func (s *Stylus) IsActive() bool {
	return s.actionState != StylusActionNone
}

func (s *Stylus) Reset() {
	if s.actionState == StylusActionDown || s.actionState == StylusActionHeld {
		s.actionState = StylusActionUp
	}
}
