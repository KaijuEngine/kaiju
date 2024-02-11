/*****************************************************************************/
/* touch.go                                                                  */
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

package hid

import "slices"

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
	Pressure float32
	X        float32
	Y        float32
	IY       float32
	State    TouchAction
	Id       int64
}

type Touch struct {
	Pointers        []*TouchPointer
	Pool            [MaxTouchPointersAvailable]TouchPointer
	Count           int
	resetTouchCount bool
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

func (p *TouchPointer) setPosition(x, y, windowHeight float32) {
	p.X = x
	p.Y = y
	p.IY = windowHeight - y
}

func (t *Touch) SetDown(id int64, x, y, windowHeight float32) {
	if p, found := t.newPointer(id); found {
		p.State = TouchActionDown
		t.Pointers = append(t.Pointers, p)
		t.SetMoved(id, x, y, windowHeight)
	}
}

func (t *Touch) SetUp(id int64, x, y, windowHeight float32) {
	if p := t.point(id); p != nil {
		p.State = TouchActionUp
		p.setPosition(x, y, windowHeight)
	}
}

func (t *Touch) SetMoved(id int64, x, y, windowHeight float32) {
	if p := t.point(id); p != nil {
		p.setPosition(x, y, windowHeight)
	}
}

func (t *Touch) SetPressure(id int64, pressure float32) {
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
	return t.Count > 0 && t.Pointers[0].State == TouchActionDown
}

func (t *Touch) Held() bool {
	if t.Count > 0 {
		s := t.Pointers[0].State
		return s == TouchActionHeld || s == TouchActionMove
	} else {
		return false
	}
}

func (t *Touch) Moved() bool {
	return t.Count > 0 && t.Pointers[0].State == TouchActionMove
}

func (t *Touch) Released() bool {
	return t.Count > 0 && t.Pointers[0].State == TouchActionUp
}

func (t *Touch) Cancelled() bool {
	return t.Count > 0 && t.Pointers[0].State == TouchActionCancel
}

func (t *Touch) SetCount(count int) {
	t.Count = count
}

func (t *Touch) Pointer(index int) *TouchPointer {
	return t.Pointers[index]
}

func (t *Touch) EndUpdate() {
	if t.resetTouchCount {
		t.Count = 0
		t.resetTouchCount = false
	}
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
			t.Count = max(0, t.Count-1)
		case TouchActionCancel:
			t.Pool[i].State = TouchActionNone
			t.Pool[i].Id = -1
			t.Pointers = t.Pointers[:0]
			t.Count = 0
		}
	}
}
