/*****************************************************************************/
/* cursor.go                                                                 */
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
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
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

import "kaiju/matrix"

type Cursor struct {
	mouse    *Mouse
	touch    *Touch
	stylus   *Stylus
	pos      matrix.Vec2
	lastPos  matrix.Vec2
	pressure float32
	distance float32
}

func NewCursor(mouse *Mouse, touch *Touch, stylus *Stylus) Cursor {
	return Cursor{
		mouse:  mouse,
		touch:  touch,
		stylus: stylus,
	}
}

func (c *Cursor) Moved() bool {
	return c.mouse.Moved() || c.touch.Moved() || c.stylus.Moved() || matrix.Vec2Approx(c.lastPos, c.pos)
}

func (c *Cursor) Pressed() bool {
	return c.mouse.Pressed(MouseButtonLeft) || c.touch.Pressed() || c.stylus.Pressed()
}

func (c *Cursor) Held() bool {
	return c.mouse.Held(MouseButtonLeft) || c.touch.Held() || c.stylus.Held()
}

func (c *Cursor) Released() bool {
	return c.mouse.Released(MouseButtonLeft) || c.touch.Released() || c.stylus.Released()
}

func (c *Cursor) Poll() {
	c.lastPos = c.pos
	c.pos = c.Position()
	if c.touch.Count > 0 {
		c.pressure = c.touch.Pointers[0].Pressure
	} else {
		c.pressure = c.stylus.Pressure
	}
	c.distance = c.stylus.Distance
}

func (c *Cursor) ScreenPosition() matrix.Vec2 {
	if c.touch.Count == 1 {
		p := c.touch.Pointer(0)
		return matrix.Vec2{p.X, p.IY}
	} else if c.stylus.IsActive() {
		return matrix.Vec2{c.stylus.X, c.stylus.IY}
	} else {
		return c.mouse.ScreenPosition()
	}
}

func (c *Cursor) UIPosition(uiSize, windowSize matrix.Vec2) matrix.Vec2 {
	wRatio := uiSize.X() / windowSize.X()
	hRatio := uiSize.Y() / windowSize.Y()
	var pos matrix.Vec2
	if c.touch.Count == 1 {
		p := c.touch.Pointer(0)
		pos = matrix.Vec2{p.X * wRatio, p.IY * hRatio}
	} else if c.stylus.IsActive() {
		pos = matrix.Vec2{c.stylus.X * wRatio, c.stylus.IY * hRatio}
	} else {
		pos = c.mouse.ScreenPosition()
	}
	return pos
}

func (c *Cursor) Position() matrix.Vec2 {
	if c.touch.Count == 1 {
		p := c.touch.Pointer(0)
		return matrix.Vec2{p.X, p.Y}
	} else if c.stylus.IsActive() {
		return matrix.Vec2{c.stylus.X, c.stylus.Y}
	} else {
		return c.mouse.Position()
	}
}
