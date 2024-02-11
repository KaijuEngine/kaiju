/*****************************************************************************/
/* stylus.go                                                                 */
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
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
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
