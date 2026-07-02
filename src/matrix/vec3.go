/******************************************************************************/
/* vec3.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"fmt"
)

const vec3StrFmt = "%f, %f, %f"

type Vec3T[T tNumber] [3]T

type Vec3 [3]Float

type Vec3MinMax struct {
	Min Vec3
	Max Vec3
}

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
func (v Vec3) AsVec4WithW(w Float) Vec4   { return Vec4{v[Vx], v[Vy], v[Vz], w} }
func (v Vec3) XYZ() (Float, Float, Float) { return v[Vx], v[Vy], v[Vz] }
func (v Vec3) XY() Vec2                   { return Vec2{v[Vx], v[Vy]} }
func (v Vec3) XZ() Vec2                   { return Vec2{v[Vx], v[Vz]} }
func (v Vec3) Width() Float               { return v[Vx] }
func (v Vec3) Height() Float              { return v[Vy] }
func (v Vec3) Depth() Float               { return v[Vz] }
func (v *Vec3) AddX(x Float)              { v[Vx] += x }
func (v *Vec3) AddY(y Float)              { v[Vy] += y }
func (v *Vec3) AddZ(z Float)              { v[Vz] += z }
func (v *Vec3) ScaleX(s Float)            { v[Vx] *= s }
func (v *Vec3) ScaleY(s Float)            { v[Vy] *= s }
func (v *Vec3) ScaleZ(s Float)            { v[Vz] *= s }

func (v Vec3) AsVec3i() Vec3i {
	return Vec3i{int32(v[Vx]), int32(v[Vy]), int32(v[Vz])}
}

func NewVec3[T1, T2, T3 tNumber](x T1, y T2, z T3) Vec3 {
	return Vec3{Float(x), Float(y), Float(z)}
}

func NewVec3XYZ[T tNumber](xyz T) Vec3 {
	return Vec3{Float(xyz), Float(xyz), Float(xyz)}
}

func Vec3FromArray[T tNumber](a [3]T) Vec3 {
	return Vec3{Float(a[0]), Float(a[1]), Float(a[2])}
}

func Vec3FromSlice[T tNumber](a []T) Vec3 {
	return Vec3{Float(a[0]), Float(a[1]), Float(a[2])}
}

func NewVec3MinMax() Vec3MinMax {
	return Vec3MinMax{
		Min: Vec3{FloatMax, FloatMax, FloatMax},
		Max: Vec3{-FloatMax, -FloatMax, -FloatMax},
	}
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

func (v Vec3) LengthSquared() Float {
	return v[Vx]*v[Vx] + v[Vy]*v[Vy] + v[Vz]*v[Vz]
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

func (v *Vec3) NegativeAssign() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
	v[Vz] = -v[Vz]
}

func (v Vec3) Inverse() Vec3 {
	return Vec3{1 / v[Vx], 1 / v[Vy], 1 / v[Vz]}
}

func (v *Vec3) InverseAssign() {
	v[Vx] = 1 / v[Vx]
	v[Vy] = 1 / v[Vy]
	v[Vz] = 1 / v[Vz]
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

func (v Vec3) Dot(other Vec3) Float {
	return Vec3Dot(v, other)
}

func (v Vec3) Cross(other Vec3) Vec3 {
	return Vec3Cross(v, other)
}

func Vec3Approx(a, b Vec3) bool {
	return Abs(a.X()-b.X()) < Tiny &&
		Abs(a.Y()-b.Y()) < Tiny &&
		Abs(a.Z()-b.Z()) < Tiny
}

func Vec3ApproxTo(a, b Vec3, delta Float) bool {
	return Abs(a.X()-b.X()) < delta &&
		Abs(a.Y()-b.Y()) < delta &&
		Abs(a.Z()-b.Z()) < delta
}

func Vec3Abs(v Vec3) Vec3 {
	return NewVec3(Abs(v.X()), Abs(v.Y()), Abs(v.Z()))
}

func Vec3Min(v ...Vec3) Vec3 {
	res := v[0]
	for i := 1; i < len(v); i++ {
		res[0] = min(res[0], v[i][0])
		res[1] = min(res[1], v[i][1])
		res[2] = min(res[2], v[i][2])
	}
	return res
}

func Vec3MinAbs(v ...Vec3) Vec3 {
	res := v[0].Abs()
	for i := 1; i < len(v); i++ {
		res[0] = min(res[0], Abs(v[i][0]))
		res[1] = min(res[1], Abs(v[i][1]))
		res[2] = min(res[2], Abs(v[i][2]))
	}
	return res
}

func Vec3Max(v ...Vec3) Vec3 {
	res := v[0]
	for i := 1; i < len(v); i++ {
		res[0] = max(res[0], v[i][0])
		res[1] = max(res[1], v[i][1])
		res[2] = max(res[2], v[i][2])
	}
	return res
}

func Vec3MaxAbs(v ...Vec3) Vec3 {
	res := v[0].Abs()
	for i := 1; i < len(v); i++ {
		res[0] = max(res[0], Abs(v[i][0]))
		res[1] = max(res[1], Abs(v[i][1]))
		res[2] = max(res[2], Abs(v[i][2]))
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
	if v.Equals(other) {
		return 0
	}
	return Acos(Vec3Dot(v, other) / (v.Length() * other.Length()))
}

// SignedAngle returns the signed angle (in radians) from v to other around the
// given axis. The sign is positive for counterclockwise rotation when looking
// along the axis direction. Assumes non-zero vectors; returns 0 if v or other
// is zero-length or they are equal. For best precision, normalize v, other, and
// axis before calling if they aren't already.
func (v Vec3) SignedAngle(other Vec3, axis Vec3) Float {
	if v.Equals(other) {
		return 0
	}
	lenV := v.Length()
	if lenV == 0 {
		return 0
	}
	lenO := other.Length()
	if lenO == 0 {
		return 0
	}
	lenA := axis.Length()
	if lenA == 0 {
		return 0
	}
	dot := Vec3Dot(v, other) / (lenV * lenO)
	cross := Vec3Cross(v, other)
	signedSin := Vec3Dot(cross, axis) / (lenV * lenO * lenA)
	return Atan2(signedSin, dot)
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

func Vec3NaN() Vec3 {
	return Vec3{NaN(), NaN(), NaN()}
}

func (v Vec3) IsZero() bool {
	return Vec3Approx(v, Vec3Zero())
}

func (v Vec3) IsInf(sign int) bool {
	return IsInf(v[Vx], sign) || IsInf(v[Vy], sign) || IsInf(v[Vz], sign)
}

func (v Vec3) IsNaN() bool {
	return IsNaN(v[Vx]) || IsNaN(v[Vy]) || IsNaN(v[Vz])
}
