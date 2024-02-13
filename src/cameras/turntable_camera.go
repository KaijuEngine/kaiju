/*****************************************************************************/
/* turntable_camera.go                                                       */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
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

package cameras

import "kaiju/matrix"

type TurntableCamera struct {
	StandardCamera
}

func ToTurntable(camera *StandardCamera) *TurntableCamera {
	tc := &TurntableCamera{
		StandardCamera: *camera,
	}
	tc.updateView = tc.internalUpdateView
	return tc
}

func (c *TurntableCamera) internalUpdateView() {
	c.view = matrix.Mat4Identity()

	tx := -c.lookAt.X()
	ty := -c.lookAt.Y()
	tz := -c.lookAt.Z()
	rx := c.pitch
	ry := c.yaw
	rz := float32(0.0)
	di := c.zoom

	a := rx * float32(0.5)
	b := ry * float32(0.5)
	cc := rz * float32(0.5)
	d := matrix.Cos(a)
	e := matrix.Sin(a)
	f := matrix.Cos(b)
	g := matrix.Sin(b)
	h := matrix.Cos(cc)
	i := matrix.Sin(cc)
	j := f*e*h + g*d*i
	k := f*-e*i + g*d*h
	l := f*d*i - g*e*h
	m := f*d*h - g*-e*i
	n := j * j
	o := k * k
	p := l * l
	q := m * m
	r := j * k
	s := k * l
	t := j * l
	u := m * j
	v := m * k
	w := m * l
	x := q + n - o - p
	y := (r + w) * 2.0
	z := (t - v) * 2.0
	A := (r - w) * 2.0
	B := q - n + o - p
	C := (s + u) * 2.0
	D := (t + v) * 2.0
	E := (s - u) * 2.0
	F := q - n - o + p
	G := di
	H := -(tx + D*G)
	I := -(ty + E*G)
	J := -(tz + F*G)
	c.view[0] = x
	c.view[1] = A
	c.view[2] = D
	c.view[3] = 0.0
	c.view[4] = y
	c.view[5] = B
	c.view[6] = E
	c.view[7] = 0.0
	c.view[8] = z
	c.view[9] = C
	c.view[10] = F
	c.view[11] = 0.0
	c.view[12] = x*H + y*I + z*J
	c.view[13] = A*H + B*I + C*J
	c.view[14] = D*H + E*I + F*J
	c.view[15] = 1.0
	c.iView = c.view
	c.iView.Inverse()
	c.updateFrustum()
}

func (c *TurntableCamera) updateViewAndPosition() {
	c.position.SetZ(c.zoom)
	c.updateView()
	c.position = c.iView.Position()
}

func (c *TurntableCamera) SetPosition(position matrix.Vec3) {
	c.position = position
	c.zoom = position.Z()
	c.updateViewAndPosition()
}

func (c *TurntableCamera) Pan(delta matrix.Vec3) {
	d := delta.Scale(c.zoom)
	u := c.Up()
	u.ScaleAssign(-d.Y())
	r := c.Right()
	r.ScaleAssign(-d.X())
	c.lookAt.AddAssign(u)
	c.lookAt.AddAssign(r)
	c.updateViewAndPosition()
}

func (c *TurntableCamera) Dolly(delta float32) {
	diff := c.position.Subtract(c.lookAt)
	length := diff.Length()
	c.zoom += delta * length
	if c.position.Z() <= 0.0 {
		c.zoom += 0.001
	}
	c.updateViewAndPosition()
}

func (c *TurntableCamera) Orbit(delta matrix.Vec3) {
	c.pitch += delta.X()
	c.yaw += delta.Y()
	c.updateViewAndPosition()
}
