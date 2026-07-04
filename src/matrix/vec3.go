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
type Vec3 = Vec3T[Float]

type Vec3MinMax struct {
	Min Vec3
	Max Vec3
}

func (v Vec3T[T]) X() T                 { return v[Vx] }
func (v Vec3T[T]) Y() T                 { return v[Vy] }
func (v Vec3T[T]) Z() T                 { return v[Vz] }
func (v *Vec3T[T]) PX() *T              { return &v[Vx] }
func (v *Vec3T[T]) PY() *T              { return &v[Vy] }
func (v *Vec3T[T]) PZ() *T              { return &v[Vz] }
func (v *Vec3T[T]) SetX(x T)            { v[Vx] = x }
func (v *Vec3T[T]) SetY(y T)            { v[Vy] = y }
func (v *Vec3T[T]) SetZ(z T)            { v[Vz] = z }
func (v Vec3T[T]) AsVec2() Vec2         { return Vec2{Float(v[Vx]), Float(v[Vy])} }
func (v Vec3T[T]) AsVec4() Vec4         { return NewVec4(v[Vx], v[Vy], v[Vz], 1) }
func (v Vec3T[T]) AsVec4WithW(w T) Vec4 { return NewVec4(v[Vx], v[Vy], v[Vz], w) }
func (v Vec3T[T]) XYZ() (T, T, T)       { return v[Vx], v[Vy], v[Vz] }
func (v Vec3T[T]) XY() Vec2             { return Vec2{Float(v[Vx]), Float(v[Vy])} }
func (v Vec3T[T]) XZ() Vec2             { return Vec2{Float(v[Vx]), Float(v[Vz])} }
func (v Vec3T[T]) Width() T             { return v[Vx] }
func (v Vec3T[T]) Height() T            { return v[Vy] }
func (v Vec3T[T]) Depth() T             { return v[Vz] }
func (v *Vec3T[T]) AddX(x T)            { v[Vx] += x }
func (v *Vec3T[T]) AddY(y T)            { v[Vy] += y }
func (v *Vec3T[T]) AddZ(z T)            { v[Vz] += z }
func (v *Vec3T[T]) ScaleX(s T)          { v[Vx] *= s }
func (v *Vec3T[T]) ScaleY(s T)          { v[Vy] *= s }
func (v *Vec3T[T]) ScaleZ(s T)          { v[Vz] *= s }

func (v Vec3T[T]) AsVec3i() Vec3i {
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

func (v Vec3T[T]) AsAligned16() [4]Float {
	return [4]Float{Float(v[Vx]), Float(v[Vy]), Float(v[Vz]), 0}
}

func (v Vec3T[T]) Add(other Vec3T[T]) Vec3T[T] {
	return Vec3T[T]{v[Vx] + other[Vx], v[Vy] + other[Vy], v[Vz] + other[Vz]}
}

func (v *Vec3T[T]) AddAssign(other Vec3T[T]) {
	v[Vx] += other[Vx]
	v[Vy] += other[Vy]
	v[Vz] += other[Vz]
}

func (v Vec3T[T]) Subtract(other Vec3T[T]) Vec3T[T] {
	return Vec3T[T]{v[Vx] - other[Vx], v[Vy] - other[Vy], v[Vz] - other[Vz]}
}

func (v *Vec3T[T]) SubtractAssign(other Vec3T[T]) {
	v[Vx] -= other[Vx]
	v[Vy] -= other[Vy]
	v[Vz] -= other[Vz]
}

func (v Vec3T[T]) Multiply(other Vec3T[T]) Vec3T[T] {
	return Vec3T[T]{v[Vx] * other[Vx], v[Vy] * other[Vy], v[Vz] * other[Vz]}
}

func (v *Vec3T[T]) MultiplyAssign(other Vec3T[T]) {
	v[Vx] *= other[Vx]
	v[Vy] *= other[Vy]
	v[Vz] *= other[Vz]
}

func (v Vec3T[T]) Divide(other Vec3T[T]) Vec3T[T] {
	return Vec3T[T]{v[Vx] / other[Vx], v[Vy] / other[Vy], v[Vz] / other[Vz]}
}

func (v *Vec3T[T]) DivideAssign(other Vec3T[T]) {
	v[Vx] /= other[Vx]
	v[Vy] /= other[Vy]
	v[Vz] /= other[Vz]
}

func (v Vec3T[T]) Scale(scalar T) Vec3T[T] {
	return Vec3T[T]{v[Vx] * scalar, v[Vy] * scalar, v[Vz] * scalar}
}

func (v *Vec3T[T]) ScaleAssign(scalar T) {
	v[Vx] *= scalar
	v[Vy] *= scalar
	v[Vz] *= scalar
}

func (v Vec3T[T]) Shrink(scalar T) Vec3T[T] {
	return Vec3T[T]{v[Vx] / scalar, v[Vy] / scalar, v[Vz] / scalar}
}

func (v *Vec3T[T]) ShrinkAssign(scalar T) {
	v[Vx] /= scalar
	v[Vy] /= scalar
	v[Vz] /= scalar
}

func (v Vec3T[T]) Length() T {
	return T(Sqrt(Vec3Dot(v, v)))
}

func (v Vec3T[T]) LengthSquared() T {
	return v[Vx]*v[Vx] + v[Vy]*v[Vy] + v[Vz]*v[Vz]
}

func (v Vec3T[T]) Normal() Vec3T[T] {
	return v.Scale(1.0 / v.Length())
}

func (v *Vec3T[T]) Normalize() {
	v.ScaleAssign(1.0 / v.Length())
}

func (v Vec3T[T]) Negative() Vec3T[T] {
	return Vec3T[T]{-v[Vx], -v[Vy], -v[Vz]}
}

func (v *Vec3T[T]) NegativeAssign() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
	v[Vz] = -v[Vz]
}

func (v Vec3T[T]) Inverse() Vec3T[T] {
	return Vec3T[T]{1 / v[Vx], 1 / v[Vy], 1 / v[Vz]}
}

func (v *Vec3T[T]) InverseAssign() {
	v[Vx] = 1 / v[Vx]
	v[Vy] = 1 / v[Vy]
	v[Vz] = 1 / v[Vz]
}

func Vec3Cross[T tNumber](v, other Vec3T[T]) Vec3T[T] {
	return Vec3T[T]{
		v[Vy]*other[Vz] - v[Vz]*other[Vy],
		v[Vz]*other[Vx] - v[Vx]*other[Vz],
		v[Vx]*other[Vy] - v[Vy]*other[Vx],
	}
}

func (v Vec3T[T]) Orthogonal() Vec3T[T] {
	other := Vec3T[T]{}
	tx, ty, tz := v[Vx], v[Vy], v[Vz]
	if tx < ty {
		if tx < tz {
			other[Vx] = T(1)
		} else {
			other[Vz] = T(0) - T(1)
		}
	} else {
		if ty < tz {
			other[Vy] = T(1)
		} else {
			other[Vz] = T(0) - T(1)
		}
	}
	return Vec3Cross(v, other)
}

func (v Vec3T[T]) Dot(other Vec3T[T]) T {
	return Vec3Dot(v, other)
}

func (v Vec3T[T]) Cross(other Vec3T[T]) Vec3T[T] {
	return Vec3Cross(v, other)
}

func Vec3Approx[T tNumber](a, b Vec3T[T]) bool {
	return Float(Abs(a.X()-b.X())) < Tiny &&
		Float(Abs(a.Y()-b.Y())) < Tiny &&
		Float(Abs(a.Z()-b.Z())) < Tiny
}

func Vec3ApproxTo[T tNumber](a, b Vec3T[T], delta T) bool {
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

func (v Vec3T[T]) Abs() Vec3T[T] {
	return Vec3T[T]{T(Abs(v[Vx])), T(Abs(v[Vy])), T(Abs(v[Vz]))}
}

func (v Vec3T[T]) Distance(other Vec3T[T]) T {
	return v.Subtract(other).Length()
}

func Vec3Dot[T tNumber](v, other Vec3T[T]) T {
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

func (v Vec3T[T]) String() string {
	return fmt.Sprintf(vec3StrFmt, Float(v[Vx]), Float(v[Vy]), Float(v[Vz]))
}

func (v Vec3T[T]) Angle(other Vec3T[T]) T {
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
func (v Vec3T[T]) SignedAngle(other Vec3T[T], axis Vec3T[T]) T {
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

func (v Vec3T[T]) Equals(other Vec3T[T]) bool {
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

func (v Vec3T[T]) LargestAxis() T {
	return max(v[Vx], v[Vy], v[Vz])
}

func (v Vec3T[T]) LargestAxisDelta() T {
	lo := min(v[Vx], v[Vy], v[Vz])
	hi := max(v[Vx], v[Vy], v[Vz])
	if Abs(lo) > Abs(hi) {
		return lo
	} else {
		return hi
	}
}

func (v Vec3T[T]) SquareDistance(b Vec3T[T]) T {
	return (v[Vx]-b[Vx])*(v[Vx]-b[Vx]) + (v[Vy]-b[Vy])*(v[Vy]-b[Vy]) + (v[Vz]-b[Vz])*(v[Vz]-b[Vz])
}

func (v Vec3T[T]) LongestAxis() int {
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

func (v Vec3T[T]) LongestAxisValue() T {
	return max(v[Vx], v[Vy], v[Vz])
}

func (v Vec3T[T]) MultiplyMat3(rhs Mat3) Vec3T[T] {
	var result Vec3T[T]
	row := rhs.RowVector(0)
	result[Vx] = T(row[Vx]*Float(v[Vx]) + row[Vy]*Float(v[Vy]) + row[Vz]*Float(v[Vz]))
	row = rhs.RowVector(1)
	result[Vy] = T(row[Vx]*Float(v[Vx]) + row[Vy]*Float(v[Vy]) + row[Vz]*Float(v[Vz]))
	row = rhs.RowVector(2)
	result[Vz] = T(row[Vx]*Float(v[Vx]) + row[Vy]*Float(v[Vy]) + row[Vz]*Float(v[Vz]))
	return result
}

func Vec3Inf(sign int) Vec3 {
	return Vec3{Inf(sign), Inf(sign), Inf(sign)}
}

func Vec3NaN() Vec3 {
	return Vec3{NaN(), NaN(), NaN()}
}

func (v Vec3T[T]) IsZero() bool {
	return Vec3Approx(v, Vec3T[T]{0, 0, 0})
}

func (v Vec3T[T]) IsInf(sign int) bool {
	return IsInf(v[Vx], sign) || IsInf(v[Vy], sign) || IsInf(v[Vz], sign)
}

func (v Vec3T[T]) IsNaN() bool {
	return IsNaN(v[Vx]) || IsNaN(v[Vy]) || IsNaN(v[Vz])
}
