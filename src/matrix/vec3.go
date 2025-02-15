/******************************************************************************/
/* vec3.go                                                                    */
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

const vec3StrFmt = "%f, %f, %f"

type Vec3 [3]Float

func (v Vec3) X() Float                   { return v[Vx] }
func (v Vec3) Y() Float                   { return v[Vy] }
func (v Vec3) Z() Float                   { return v[Vz] }
func (v *Vec3) PX() *Float                { return &v[Vx] }
func (v *Vec3) PY() *Float                { return &v[Vy] }
func (v *Vec3) PZ() *Float                { return &v[Vz] }
func (v *Vec3) SetX(x Float)              { v[Vx] = x }
func (v *Vec3) SetY(y Float)              { v[Vy] = y }
func (v *Vec3) SetZ(z Float)              { v[Vz] = z }
func (v Vec3) AsVec2() Vec2               { return Vec2(v[:Vz]) }
func (v Vec3) AsVec4() Vec4               { return Vec4{v[Vx], v[Vy], v[Vz], 1} }
func (v Vec3) XYZ() (Float, Float, Float) { return v[Vx], v[Vy], v[Vz] }

func (v Vec3) AsVec3i() Vec3i {
	return Vec3i{int32(v[Vx]), int32(v[Vy]), int32(v[Vz])}
}

func NewVec3(x, y, z Float) Vec3 {
	return Vec3{x, y, z}
}

func Vec3FromArray(a [3]Float) Vec3 {
	return Vec3{a[0], a[1], a[2]}
}

func Vec3FromSlice(a []Float) Vec3 {
	return Vec3{a[0], a[1], a[2]}
}

func (v Vec3) AsAligned16() [4]Float {
	return [4]Float{v[Vx], v[Vy], v[Vz], 0}
}

func (v Vec3) Add(other Vec3) Vec3 {
	return Vec3{v[Vx] + other[Vx], v[Vy] + other[Vy], v[Vz] + other[Vz]}
}

func (v *Vec3) AddAssign(other Vec3) {
	v[Vx] += other[Vx]
	v[Vy] += other[Vy]
	v[Vz] += other[Vz]
}

func (v Vec3) Subtract(other Vec3) Vec3 {
	return Vec3{v[Vx] - other[Vx], v[Vy] - other[Vy], v[Vz] - other[Vz]}
}

func (v *Vec3) SubtractAssign(other Vec3) {
	v[Vx] -= other[Vx]
	v[Vy] -= other[Vy]
	v[Vz] -= other[Vz]
}

func (v Vec3) Multiply(other Vec3) Vec3 {
	return Vec3{v[Vx] * other[Vx], v[Vy] * other[Vy], v[Vz] * other[Vz]}
}

func (v *Vec3) MultiplyAssign(other Vec3) {
	v[Vx] *= other[Vx]
	v[Vy] *= other[Vy]
	v[Vz] *= other[Vz]
}

func (v Vec3) Divide(other Vec3) Vec3 {
	return Vec3{v[Vx] / other[Vx], v[Vy] / other[Vy], v[Vz] / other[Vz]}
}

func (v *Vec3) DivideAssign(other Vec3) {
	v[Vx] /= other[Vx]
	v[Vy] /= other[Vy]
	v[Vz] /= other[Vz]
}

func (v Vec3) Scale(scalar Float) Vec3 {
	return Vec3{v[Vx] * scalar, v[Vy] * scalar, v[Vz] * scalar}
}

func (v *Vec3) ScaleAssign(scalar Float) {
	v[Vx] *= scalar
	v[Vy] *= scalar
	v[Vz] *= scalar
}

func (v Vec3) Shrink(scalar Float) Vec3 {
	return Vec3{v[Vx] / scalar, v[Vy] / scalar, v[Vz] / scalar}
}

func (v *Vec3) ShrinkAssign(scalar Float) {
	v[Vx] /= scalar
	v[Vy] /= scalar
	v[Vz] /= scalar
}

func (v Vec3) Length() Float {
	return Sqrt(Vec3Dot(v, v))
}

func (v Vec3) Normal() Vec3 {
	return v.Scale(1.0 / v.Length())
}

func (v *Vec3) Normalize() {
	v.ScaleAssign(1.0 / v.Length())
}

func (v Vec3) Negative() Vec3 {
	return Vec3{-v[Vx], -v[Vy], -v[Vz]}
}

func (v *Vec3) Inverse() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
	v[Vz] = -v[Vz]
}

func Vec3Cross(v, other Vec3) Vec3 {
	return Vec3{
		v[Vy]*other[Vz] - v[Vz]*other[Vy],
		v[Vz]*other[Vx] - v[Vx]*other[Vz],
		v[Vx]*other[Vy] - v[Vy]*other[Vx],
	}
}

func (v Vec3) Orthogonal() Vec3 {
	tx := v.X()
	ty := v.Y()
	tz := v.Z()
	var other Vec3
	if tx < ty {
		if tx < tz {
			other = Vec3Right()
		} else {
			other = Vec3Forward()
		}
	} else {
		if ty < tz {
			other = Vec3Up()
		} else {
			other = Vec3Forward()
		}
	}
	return Vec3Cross(v, other)
}

func Vec3Approx(a, b Vec3) bool {
	return Abs(a.X()-b.X()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Y()-b.Y()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Z()-b.Z()) < math.SmallestNonzeroFloat32
}

func Vec3ApproxTo(a, b Vec3, delta Float) bool {
	return Abs(a.X()-b.X()) < delta &&
		Abs(a.Y()-b.Y()) < delta &&
		Abs(a.Z()-b.Z()) < delta
}

func Vec3Min(v ...Vec3) Vec3 {
	res := v[0]
	for i := 1; i < len(v); i++ {
		res[0] = Min(res[0], v[i][0])
		res[1] = Min(res[1], v[i][1])
		res[2] = Min(res[2], v[i][2])
	}
	return res
}

func Vec3MinAbs(v ...Vec3) Vec3 {
	res := v[0].Abs()
	for i := 1; i < len(v); i++ {
		res[0] = Min(res[0], Abs(v[i][0]))
		res[1] = Min(res[1], Abs(v[i][1]))
		res[2] = Min(res[2], Abs(v[i][2]))
	}
	return res
}

func Vec3Max(v ...Vec3) Vec3 {
	res := v[0]
	for i := 1; i < len(v); i++ {
		res[0] = Max(res[0], v[i][0])
		res[1] = Max(res[1], v[i][1])
		res[2] = Max(res[2], v[i][2])
	}
	return res
}

func Vec3MaxAbs(v ...Vec3) Vec3 {
	res := v[0].Abs()
	for i := 1; i < len(v); i++ {
		res[0] = Max(res[0], Abs(v[i][0]))
		res[1] = Max(res[1], Abs(v[i][1]))
		res[2] = Max(res[2], Abs(v[i][2]))
	}
	return res
}

func (v Vec3) Abs() Vec3 {
	return Vec3{Abs(v[Vx]), Abs(v[Vy]), Abs(v[Vz])}
}

func (v Vec3) Distance(other Vec3) Float {
	return v.Subtract(other).Length()
}

func Vec3Dot(v, other Vec3) Float {
	return v[Vx]*other[Vx] + v[Vy]*other[Vy] + v[Vz]*other[Vz]
}

func Vec3Lerp(from, to Vec3, t Float) Vec3 {
	return from.Add(to.Subtract(from).Scale(t))
}

func Vec3FromString(str string) Vec3 {
	var v Vec3
	fmt.Sscanf(str, vec3StrFmt, &v[Vx], &v[Vy], &v[Vz])
	return v
}

func (v Vec3) String() string {
	return fmt.Sprintf(vec3StrFmt, v[Vx], v[Vy], v[Vz])
}

func (v Vec3) Angle(other Vec3) Float {
	return Acos(Vec3Dot(v, other) / (v.Length() * other.Length()))
}

func (v Vec3) Equals(other Vec3) bool {
	return Vec3Approx(v, other)
}

func Vec3Up() Vec3       { return Vec3{0, 1, 0} }
func Vec3Down() Vec3     { return Vec3{0, -1, 0} }
func Vec3Left() Vec3     { return Vec3{-1, 0, 0} }
func Vec3Right() Vec3    { return Vec3{1, 0, 0} }
func Vec3Forward() Vec3  { return Vec3{0, 0, -1} }
func Vec3Backward() Vec3 { return Vec3{0, 0, 1} }
func Vec3Zero() Vec3     { return Vec3{0, 0, 0} }
func Vec3One() Vec3      { return Vec3{1, 1, 1} }
func Vec3Half() Vec3     { return Vec3{0.5, 0.5, 0.5} }
func Vec3Largest() Vec3  { return Vec3{FloatMax, FloatMax, FloatMax} }

func (v Vec3) LargestAxis() Float {
	return max(v[Vx], v[Vy], v[Vz])
}

func (v Vec3) LargestAxisDelta() Float {
	lo := min(v[Vx], v[Vy], v[Vz])
	hi := max(v[Vx], v[Vy], v[Vz])
	if Abs(lo) > Abs(hi) {
		return lo
	} else {
		return hi
	}
}

func (v Vec3) SquareDistance(b Vec3) Float {
	return (v[Vx]-b[Vx])*(v[Vx]-b[Vx]) + (v[Vy]-b[Vy])*(v[Vy]-b[Vy]) + (v[Vz]-b[Vz])*(v[Vz]-b[Vz])
}

func (v Vec3) LongestAxis() int {
	if v[Vx] > v[Vy] {
		if v[Vx] > v[Vz] {
			return Vx
		}
		return Vz
	}
	if v[Vy] > v[Vz] {
		return Vy
	}
	return Vz
}

func (v Vec3) LongestAxisValue() Float {
	return max(v[Vx], v[Vy], v[Vz])
}

func (v Vec3) MultiplyMat3(rhs Mat3) Vec3 {
	var result Vec3
	row := rhs.RowVector(0)
	result[Vx] = Vec3Dot(row, v)
	row = rhs.RowVector(1)
	result[Vy] = Vec3Dot(row, v)
	row = rhs.RowVector(2)
	result[Vz] = Vec3Dot(row, v)
	return result
}

func Vec3Inf(sign int) Vec3 {
	return Vec3{Inf(sign), Inf(sign), Inf(sign)}
}

func (v Vec3) IsZero() bool {
	return Vec3Approx(v, Vec3Zero())
}
