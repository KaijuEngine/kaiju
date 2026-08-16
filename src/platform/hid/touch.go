/******************************************************************************/
/* touch.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package hid

import (
	"slices"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
)

const (
	MaxTouchPointersAvailable = 10
)

type TouchAction = int

// TODO:  Move this to platform specific code
const (
	internalTouchActionDown TouchAction = iota
	internalTouchActionUp
	internalTouchActionMove
)

const (
	TouchActionNone   TouchAction = -1
	TouchActionDown   TouchAction = internalTouchActionDown
	TouchActionUp     TouchAction = internalTouchActionUp
	TouchActionMove   TouchAction = internalTouchActionMove
	TouchActionCancel TouchAction = -2
	TouchActionHeld   TouchAction = -3
)

type TouchPointer struct {
	Pressure matrix.Float
	X        matrix.Float
	Y        matrix.Float
	SX       matrix.Float
	SY       matrix.Float
	State    TouchAction
	Id       int64
}

type Touch struct {
	Pointers []*TouchPointer
	Pool     [MaxTouchPointersAvailable]TouchPointer
}

func NewTouch() Touch {
	t := Touch{
		Pointers: make([]*TouchPointer, 0, MaxTouchPointersAvailable),
	}
	for i := 0; i < MaxTouchPointersAvailable; i++ {
		t.Pool[i].Id = -1
		t.Pool[i].State = TouchActionNone
	}
	return t
}

func (t *Touch) newPointer(id int64) (*TouchPointer, bool) {
	var p *TouchPointer
	found := false
	for i := 0; i < MaxTouchPointersAvailable && !found; i++ {
		if t.Pool[i].Id == -1 {
			p = &t.Pool[i]
			p.Id = id
			found = true
		}
	}
	return p, found
}

func (t *Touch) point(id int64) *TouchPointer {
	var p *TouchPointer
	for i := 0; i < MaxTouchPointersAvailable && p == nil; i++ {
		if t.Pool[i].Id == id {
			p = &t.Pool[i]
		}
	}
	return p
}

func (p *TouchPointer) setPosition(x, y, windowHeight matrix.Float) {
	p.X = x
	p.Y = windowHeight - y
	p.SX = x
	p.SY = y
}

func (t *Touch) SetDown(id int64, x, y, windowHeight matrix.Float) {
	if p, found := t.newPointer(id); found {
		p.State = TouchActionDown
		t.Pointers = append(t.Pointers, p)
		t.SetMoved(id, x, y, windowHeight)
	}
}

func (t *Touch) SetUp(id int64, x, y, windowHeight matrix.Float) {
	if p := t.point(id); p != nil {
		p.State = TouchActionUp
		p.setPosition(x, y, windowHeight)
	}
}

func (t *Touch) SetMoved(id int64, x, y, windowHeight matrix.Float) {
	if p := t.point(id); p != nil {
		p.setPosition(x, y, windowHeight)
	} else {
		t.SetDown(id, x, y, windowHeight)
	}
}

func (t *Touch) SetPressure(id int64, pressure matrix.Float) {
	if p := t.point(id); p != nil {
		p.Pressure = pressure
	}
}

func (t *Touch) Cancel() {
	for i := 0; i < MaxTouchPointersAvailable; i++ {
		t.Pool[i].State = TouchActionCancel
	}
}

func (t *Touch) Pressed() bool {
	return len(t.Pointers) > 0 && t.Pointers[0].State == TouchActionDown
}

func (t *Touch) Held() bool {
	if len(t.Pointers) > 0 {
		s := t.Pointers[0].State
		return s == TouchActionHeld || s == TouchActionMove
	} else {
		return false
	}
}

func (t *Touch) Moved() bool {
	return len(t.Pointers) > 0 && t.Pointers[0].State == TouchActionMove
}

func (t *Touch) Released() bool {
	return len(t.Pointers) > 0 && t.Pointers[0].State == TouchActionUp
}

func (t *Touch) Cancelled() bool {
	return len(t.Pointers) > 0 && t.Pointers[0].State == TouchActionCancel
}

func (t *Touch) Pointer(index int) *TouchPointer {
	return t.Pointers[index]
}

func (t *Touch) EndUpdate() {
	for i := MaxTouchPointersAvailable - 1; i >= 0; i-- {
		switch t.Pool[i].State {
		case TouchActionDown:
			fallthrough
		case TouchActionMove:
			t.Pool[i].State = TouchActionHeld
		case TouchActionUp:
			t.Pool[i].State = TouchActionNone
			t.Pool[i].Id = -1
			t.Pointers = slices.Delete(t.Pointers, i, i+1)
		case TouchActionCancel:
			t.Pool[i].State = TouchActionNone
			t.Pool[i].Id = -1
			t.Pointers = klib.WipeSlice(t.Pointers)
		}
	}
}

func (t *Touch) Reset() {
	for i := 0; i < MaxTouchPointersAvailable; i++ {
		if t.Pool[i].State == TouchActionDown || t.Pool[i].State == TouchActionHeld {
			t.Pool[i].State = TouchActionUp
		}
	}
}
