/******************************************************************************/
/* vec2.go                                                                    */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package matrix

import (
	"fmt"
	"math"
)

const vec2StrFmt = "%f, %f"

type Vec2 [2]Float

func (v Vec2) X() Float           { return v[Vx] }
func (v Vec2) Y() Float           { return v[Vy] }
func (v Vec2) Width() Float       { return v[Vx] }
func (v Vec2) Height() Float      { return v[Vy] }
func (v *Vec2) PX() *Float        { return &v[Vx] }
func (v *Vec2) PY() *Float        { return &v[Vy] }
func (v *Vec2) SetX(x Float)      { v[Vx] = x }
func (v *Vec2) SetY(y Float)      { v[Vy] = y }
func (v *Vec2) SetWidth(x Float)  { v[Vx] = x }
func (v *Vec2) SetHeight(y Float) { v[Vy] = y }
func (v *Vec2) AsVec3() Vec3      { return NewVec3(v[Vx], v[Vy], 0) }
func (v Vec2) XY() (Float, Float) { return v[Vx], v[Vy] }

func (v Vec2) AsVec2i() Vec2i {
	return Vec2i{int32(v[Vx]), int32(v[Vy])}
}

func NewVec2(x, y Float) Vec2 {
	return Vec2{x, y}
}

func Vec2FromArray(a [2]Float) Vec2 {
	return Vec2{a[0], a[1]}
}

func Vec2FromSlice(a []Float) Vec2 {
	return Vec2{a[0], a[1]}
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{v[Vx] + other[Vx], v[Vy] + other[Vy]}
}

func (v *Vec2) AddAssign(other Vec2) {
	v[Vx] += other[Vx]
	v[Vy] += other[Vy]
}

func (v Vec2) Subtract(other Vec2) Vec2 {
	return Vec2{v[Vx] - other[Vx], v[Vy] - other[Vy]}
}

func (v *Vec2) SubtractAssign(other Vec2) {
	v[Vx] -= other[Vx]
	v[Vy] -= other[Vy]
}

func (v Vec2) Multiply(other Vec2) Vec2 {
	return Vec2{v[Vx] * other[Vx], v[Vy] * other[Vy]}
}

func (v *Vec2) MultiplyAssign(other Vec2) {
	v[Vx] *= other[Vx]
	v[Vy] *= other[Vy]
}

func (v Vec2) Divide(other Vec2) Vec2 {
	return Vec2{v[Vx] / other[Vx], v[Vy] / other[Vy]}
}

func (v *Vec2) DivideAssign(other Vec2) {
	v[Vx] /= other[Vx]
	v[Vy] /= other[Vy]
}

func (v Vec2) Scale(scalar Float) Vec2 {
	return Vec2{v[Vx] * scalar, v[Vy] * scalar}
}

func (v *Vec2) ScaleAssign(scalar Float) {
	v[Vx] *= scalar
	v[Vy] *= scalar
}

func (v Vec2) Shrink(scalar Float) Vec2 {
	return Vec2{v[Vx] / scalar, v[Vy] / scalar}
}

func (v *Vec2) ShrinkAssign(scalar Float) {
	v[Vx] /= scalar
	v[Vy] /= scalar
}

func (v Vec2) Length() Float {
	return Sqrt(Vec2Dot(v, v))
}

func (v Vec2) Normal() Vec2 {
	return v.Scale(1.0 / v.Length())
}

func (v *Vec2) Normalize() {
	v.ScaleAssign(1.0 / v.Length())
}

func (v Vec2) Negative() Vec2 {
	return Vec2{-v[Vx], -v[Vy]}
}

func (v *Vec2) Inverse() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
}

func Vec2Roughly(a, b Vec2) bool {
	return Abs(a.X()-b.X()) < Roughly &&
		Abs(a.Y()-b.Y()) < Roughly
}

func Vec2Nearly(a, b Vec2) bool {
	return Abs(a.X()-b.X()) < Tiny &&
		Abs(a.Y()-b.Y()) < Tiny
}

func Vec2Approx(a, b Vec2) bool {
	return Abs(a.X()-b.X()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Y()-b.Y()) < math.SmallestNonzeroFloat32
}

func Vec2ApproxTo(a, b Vec2, delta Float) bool {
	return Abs(a.X()-b.X()) < delta && Abs(a.Y()-b.Y()) < delta
}

func Vec2Min(a, b Vec2) Vec2 {
	return Vec2{
		Min(a[Vx], b[Vx]),
		Min(a[Vy], b[Vy]),
	}
}

func Vec2MinAbs(a, b Vec2) Vec2 {
	return Vec2{
		Min(Abs(a[Vx]), Abs(b[Vx])),
		Min(Abs(a[Vy]), Abs(b[Vy])),
	}
}

func Vec2Max(a, b Vec2) Vec2 {
	return Vec2{
		Max(a[Vx], b[Vx]),
		Max(a[Vy], b[Vy]),
	}
}

func Vec2MaxAbs(a, b Vec2) Vec2 {
	return Vec2{
		Max(Abs(a[Vx]), Abs(b[Vx])),
		Max(Abs(a[Vy]), Abs(b[Vy])),
	}
}

func (v Vec2) Abs() Vec2 {
	return Vec2{Abs(v[Vx]), Abs(v[Vy])}
}

func (v Vec2) Distance(other Vec2) Float {
	return v.Subtract(other).Length()
}

func Vec2Dot(v, other Vec2) Float {
	return v[Vx]*other[Vx] + v[Vy]*other[Vy]
}

func Vec2Lerp(from, to Vec2, t Float) Vec2 {
	return from.Add(to.Subtract(from).Scale(t))
}

func Vec2FromString(str string) Vec2 {
	var v Vec2
	fmt.Sscanf(str, vec2StrFmt, &v[Vx], &v[Vy])
	return v
}

func (v Vec2) String() string {
	return fmt.Sprintf(vec2StrFmt, v[Vx], v[Vy])
}

func (v Vec2) Angle(other Vec2) Float {
	return Acos(Vec2Dot(v, other) / (v.Length() * other.Length()))
}

func (v Vec2) Equals(other Vec2) bool {
	return Vec2Approx(v, other)
}

func Vec2Up() Vec2      { return Vec2{0, 1} }
func Vec2Down() Vec2    { return Vec2{0, -1} }
func Vec2Left() Vec2    { return Vec2{-1, 0} }
func Vec2Right() Vec2   { return Vec2{1, 0} }
func Vec2Zero() Vec2    { return Vec2{0, 0} }
func Vec2One() Vec2     { return Vec2{1, 1} }
func Vec2Half() Vec2    { return Vec2{0.5, 0.5} }
func Vec2Largest() Vec2 { return Vec2{FloatMax, FloatMax} }

func (v Vec2) LargestAxis() Float {
	return max(v[Vx], v[Vy])
}

func (v Vec2) LargestAxisDelta() Float {
	lo := min(v[Vx], v[Vy])
	hi := max(v[Vx], v[Vy])
	if Abs(lo) > Abs(hi) {
		return lo
	} else {
		return hi
	}
}
