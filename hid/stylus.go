package hid

type StylusActionState = int

// TODO:  This is android specific stuff
const (
	AMotionEventActionIdle = 99
	AMotionEventActionHeld = 100

	AMotionEventActionHoverEnter = 1 + iota
	AMotionEventActionHoverMove
	AMotionEventActionHoverExit
	AMotionEventActionHover
	AMotionEventActionDown
	AMotionEventActionMove
	AMotionEventActionUp
)

const (
	StylusActionNone       StylusActionState = AMotionEventActionIdle
	StylusActionHoverEnter StylusActionState = AMotionEventActionHoverEnter
	StylusActionHoverMove  StylusActionState = AMotionEventActionHoverMove
	StylusActionHoverExit  StylusActionState = AMotionEventActionHoverExit
	StylusActionDown       StylusActionState = AMotionEventActionDown
	StylusActionMove       StylusActionState = AMotionEventActionMove
	StylusActionUp         StylusActionState = AMotionEventActionUp
	StylusActionHeld       StylusActionState = AMotionEventActionHeld
	StylusActionHover      StylusActionState = AMotionEventActionHover
)

type Stylus struct {
	X           float32
	Y           float32
	IY          float32
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

func (s *Stylus) SetDistance(distance float32) {
	s.Distance = distance
}

func (s *Stylus) Set(x, y, windowHeight, pressure float32) {
	s.X = x
	s.Y = y
	s.IY = windowHeight - y
	s.Pressure = pressure
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
