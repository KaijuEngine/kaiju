/******************************************************************************/
/* plane.go                                                                   */
/******************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/******************************************************************************/
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
/******************************************************************************/

package collision

import "kaiju/matrix"

type Plane struct {
	Normal matrix.Vec3
	Dot    float32
}

func PlaneCCW(a, b, c matrix.Vec3) Plane {
	var p Plane
	e0 := b.Subtract(a)
	e1 := c.Subtract(a)
	p.Normal = matrix.Vec3Cross(e0, e1).Normal()
	p.Dot = matrix.Vec3Dot(p.Normal, a)
	return p
}

func (p *Plane) SetFloatValue(value float32, index int) {
	switch index {
	case 0:
		p.Normal.SetX(value)
	case 1:
		p.Normal.SetY(value)
	case 2:
		p.Normal.SetZ(value)
	case 3:
		p.Dot = value
	}
}

func (p Plane) ToArray() [4]float32 {
	return [4]float32{p.Normal.X(), p.Normal.Y(), p.Normal.Z(), p.Dot}
}

func (p Plane) ToVec4() matrix.Vec4 {
	return matrix.Vec4{p.Normal.X(), p.Normal.Y(), p.Normal.Z(), p.Dot}
}

func (p Plane) ClosestPoint(point matrix.Vec3) matrix.Vec3 {
	// If normalized, t := matrix.Vec3Dot(point, p.Normal) - p.Dot
	t := (matrix.Vec3Dot(p.Normal, point) - p.Dot) / matrix.Vec3Dot(p.Normal, p.Normal)
	return point.Subtract(p.Normal.Scale(t))
}

func (p Plane) Distance(point matrix.Vec3) float32 {
	// If normalized, return matrix.Vec3Dot(p.Normal, point) - p.Dot
	return (matrix.Vec3Dot(p.Normal, point) - p.Dot) / matrix.Vec3Dot(p.Normal, p.Normal)
}

func PointOutsideOfPlane(p, a, b, c, d matrix.Vec3) bool {
	signp := matrix.Vec3Dot(p.Subtract(a), matrix.Vec3Cross(b.Subtract(a), c.Subtract(a)))
	signd := matrix.Vec3Dot(d.Subtract(a), matrix.Vec3Cross(b.Subtract(a), c.Subtract(a)))
	return signp*signd < 0
}
