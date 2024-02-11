/*****************************************************************************/
/* vec4.go                                                                   */
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

package matrix

import (
	"fmt"
	"math"
)

const vec4StrFmt = "%f, %f, %f, %f"

type Vec4 [4]Float

func (v Vec4) X() Float                           { return v[Vx] }
func (v Vec4) Y() Float                           { return v[Vy] }
func (v Vec4) Z() Float                           { return v[Vz] }
func (v Vec4) W() Float                           { return v[Vw] }
func (v Vec4) Left() Float                        { return v[Vx] }
func (v Vec4) Top() Float                         { return v[Vy] }
func (v Vec4) Right() Float                       { return v[Vz] }
func (v Vec4) Bottom() Float                      { return v[Vw] }
func (v Vec4) Width() Float                       { return v[Vz] }
func (v Vec4) Height() Float                      { return v[Vw] }
func (v *Vec4) PX() *Float                        { return &v[Vx] }
func (v *Vec4) PY() *Float                        { return &v[Vy] }
func (v *Vec4) PZ() *Float                        { return &v[Vz] }
func (v *Vec4) PW() *Float                        { return &v[Vw] }
func (v *Vec4) SetX(x Float)                      { v[Vx] = x }
func (v *Vec4) SetY(y Float)                      { v[Vy] = y }
func (v *Vec4) SetZ(z Float)                      { v[Vz] = z }
func (v *Vec4) SetW(w Float)                      { v[Vw] = w }
func (v *Vec4) SetLeft(x Float)                   { v[Vx] = x }
func (v *Vec4) SetTop(y Float)                    { v[Vy] = y }
func (v *Vec4) SetRight(z Float)                  { v[Vz] = z }
func (v *Vec4) SetBottom(w Float)                 { v[Vw] = w }
func (v *Vec4) SetWidth(z Float)                  { v[Vz] = z }
func (v *Vec4) SetHeight(w Float)                 { v[Vw] = w }
func (v Vec4) AsVec3() Vec3                       { return Vec3(v[:Vw]) }
func (v Vec4) XYZW() (Float, Float, Float, Float) { return v[Vx], v[Vy], v[Vz], v[Vw] }

func (v Vec4) AsVec4i() Vec4i {
	return Vec4i{int32(v[Vx]), int32(v[Vy]), int32(v[Vz]), int32(v[Vw])}
}

func NewVec4(x, y, z, w Float) Vec4 {
	return Vec4{x, y, z, w}
}

func Vec4FromArray(a [4]Float) Vec4 {
	return Vec4{a[0], a[1], a[2], a[3]}
}

func Vec4FromSlice(a []Float) Vec4 {
	return Vec4{a[0], a[1], a[2], a[3]}
}

func (v Vec4) Add(other Vec4) Vec4 {
	return Vec4{v[Vx] + other[Vx], v[Vy] + other[Vy], v[Vz] + other[Vz], v[Vw] + other[Vw]}
}

func (v *Vec4) AddAssign(other Vec4) {
	v[Vx] += other[Vx]
	v[Vy] += other[Vy]
	v[Vz] += other[Vz]
	v[Vw] += other[Vw]
}

func (v Vec4) Subtract(other Vec4) Vec4 {
	return Vec4{v[Vx] - other[Vx], v[Vy] - other[Vy], v[Vz] - other[Vz], v[Vw] - other[Vw]}
}

func (v *Vec4) SubtractAssign(other Vec4) {
	v[Vx] -= other[Vx]
	v[Vy] -= other[Vy]
	v[Vz] -= other[Vz]
	v[Vw] -= other[Vw]
}

func (v Vec4) Multiply(other Vec4) Vec4 {
	return Vec4{v[Vx] * other[Vx], v[Vy] * other[Vy], v[Vz] * other[Vz], v[Vw] * other[Vw]}
}

func (v *Vec4) MultiplyAssign(other Vec4) {
	v[Vx] *= other[Vx]
	v[Vy] *= other[Vy]
	v[Vz] *= other[Vz]
	v[Vw] *= other[Vw]
}

func (v Vec4) Divide(other Vec4) Vec4 {
	return Vec4{v[Vx] / other[Vx], v[Vy] / other[Vy], v[Vz] / other[Vz], v[Vw] / other[Vw]}
}

func (v *Vec4) DivideAssign(other Vec4) {
	v[Vx] /= other[Vx]
	v[Vy] /= other[Vy]
	v[Vz] /= other[Vz]
	v[Vw] /= other[Vw]
}

func (v Vec4) Scale(scalar Float) Vec4 {
	return Vec4{v[Vx] * scalar, v[Vy] * scalar, v[Vz] * scalar, v[Vw] * scalar}
}

func (v *Vec4) ScaleAssign(scalar Float) {
	v[Vx] *= scalar
	v[Vy] *= scalar
	v[Vz] *= scalar
	v[Vw] *= scalar
}

func (v Vec4) Shrink(scalar Float) Vec4 {
	return Vec4{v[Vx] / scalar, v[Vy] / scalar, v[Vz] / scalar, v[Vw] / scalar}
}

func (v *Vec4) ShrinkAssign(scalar Float) {
	v[Vx] /= scalar
	v[Vy] /= scalar
	v[Vz] /= scalar
	v[Vw] /= scalar
}

func (v Vec4) Length() Float {
	return Sqrt(Vec4Dot(v, v))
}

func (v Vec4) Normal() Vec4 {
	return v.Scale(1.0 / v.Length())
}

func (v *Vec4) Normalize() {
	v.ScaleAssign(1.0 / v.Length())
}

func (v Vec4) Negative() Vec4 {
	return Vec4{-v[Vx], -v[Vy], -v[Vz], -v[Vw]}
}

func (v *Vec4) Inverse() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
	v[Vz] = -v[Vz]
	v[Vw] = -v[Vw]
}

func Vec4Approx(a, b Vec4) bool {
	return Abs(a.X()-b.X()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Y()-b.Y()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Z()-b.Z()) < math.SmallestNonzeroFloat32 &&
		Abs(a.W()-b.W()) < math.SmallestNonzeroFloat32
}

func Vec4ApproxTo(a, b Vec4, delta Float) bool {
	return Abs(a.X()-b.X()) < delta &&
		Abs(a.Y()-b.Y()) < delta &&
		Abs(a.Z()-b.Z()) < delta &&
		Abs(a.W()-b.W()) < delta
}

func Vec4Min(a, b Vec4) Vec4 {
	return Vec4{
		Min(a[Vx], b[Vx]),
		Min(a[Vy], b[Vy]),
		Min(a[Vz], b[Vz]),
		Min(a[Vw], b[Vw]),
	}
}

func Vec4MinAbs(a, b Vec4) Vec4 {
	return Vec4{
		Min(Abs(a[Vx]), Abs(b[Vx])),
		Min(Abs(a[Vy]), Abs(b[Vy])),
		Min(Abs(a[Vz]), Abs(b[Vz])),
		Min(Abs(a[Vw]), Abs(b[Vw])),
	}
}

func Vec4Max(a, b Vec4) Vec4 {
	return Vec4{
		Max(a[Vx], b[Vx]),
		Max(a[Vy], b[Vy]),
		Max(a[Vz], b[Vz]),
		Max(a[Vw], b[Vw]),
	}
}

func Vec4MaxAbs(a, b Vec4) Vec4 {
	return Vec4{
		Max(Abs(a[Vx]), Abs(b[Vx])),
		Max(Abs(a[Vy]), Abs(b[Vy])),
		Max(Abs(a[Vz]), Abs(b[Vz])),
		Max(Abs(a[Vw]), Abs(b[Vw])),
	}
}

func (v Vec4) Abs() Vec4 {
	return Vec4{Abs(v[Vx]), Abs(v[Vy]), Abs(v[Vz]), Abs(v[Vw])}
}

func (v Vec4) Distance(other Vec4) Float {
	return v.Subtract(other).Length()
}

func Vec4Dot(v, other Vec4) Float {
	return v[Vx]*other[Vx] + v[Vy]*other[Vy] + v[Vz]*other[Vz] + v[Vw]*other[Vw]
}

func Vec4Lerp(from, to Vec4, t Float) Vec4 {
	return from.Add(to.Subtract(from).Scale(t))
}

func Vec4FromString(str string) Vec4 {
	var v Vec4
	fmt.Sscanf(str, vec4StrFmt, &v[Vx], &v[Vy], &v[Vz], &v[Vw])
	return v
}

func (v Vec4) String() string {
	return fmt.Sprintf(vec4StrFmt, v[Vx], v[Vy], v[Vz], v[Vw])
}

func (v Vec4) Angle(other Vec4) Float {
	return Acos(Vec4Dot(v, other) / (v.Length() * other.Length()))
}

func (v Vec4) Equals(other Vec4) bool {
	return Vec4Approx(v, other)
}

func Vec4Zero() Vec4 { return Vec4{0, 0, 0, 0} }
func Vec4One() Vec4  { return Vec4{1, 1, 1, 1} }
func Vec4Half() Vec4 { return Vec4{0.5, 0.5, 0.5, 0.5} }
func Vec4Largest() Vec4 {
	return Vec4{FloatMax, FloatMax, FloatMax, FloatMax}
}

func (v Vec4) LargestAxis() Float {
	return max(v[Vx], v[Vy], v[Vz], v[Vw])
}

func (v Vec4) MultiplyMat4(rhs Mat4) Vec4 {
	var result Vec4
	row := rhs.RowVector(0)
	result[Vx] = Vec4Dot(row, v)
	row = rhs.RowVector(1)
	result[Vy] = Vec4Dot(row, v)
	row = rhs.RowVector(2)
	result[Vz] = Vec4Dot(row, v)
	row = rhs.RowVector(3)
	result[Vw] = Vec4Dot(row, v)
	return result
}

func (v Vec4) BoxContains(x, y Float) bool {
	return v.X() <= x && v.X()+v.Width() >= x && v.Y() <= y && v.Y()+v.Height() >= y
}

func (v Vec4) AreaContains(x, y Float) bool {
	return v.X() <= x && v.Right() >= x && v.Y() >= y && v.Bottom() <= y
}

func (v Vec4) ScreenAreaContains(x, y Float) bool {
	return v.X() <= x && v.Right() >= x && v.Y() <= y && v.Bottom() >= y
}
