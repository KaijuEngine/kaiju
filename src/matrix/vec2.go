/******************************************************************************/
/* vec2.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"fmt"
)

const vec2StrFmt = "%f, %f"

type Vec2T[T tNumber] [2]T
type Vec2 = Vec2T[Float]

func NewVec2[T1, T2 tNumber](x T1, y T2) Vec2 {
	return Vec2{Float(x), Float(y)}
}

func Vec2FromArray[T tNumber](a [2]T) Vec2 {
	return Vec2{Float(a[0]), Float(a[1])}
}

func Vec2FromSlice[T tNumber](a []T) Vec2 {
	return Vec2{Float(a[0]), Float(a[1])}
}

func Vec2Roughly(a, b Vec2) bool {
	return Abs(a.X()-b.X()) < Roughly &&
		Abs(a.Y()-b.Y()) < Roughly
}

func Vec2Nearly(a, b Vec2) bool {
	return Abs(a.X()-b.X()) < Tiny &&
		Abs(a.Y()-b.Y()) < Tiny
}

func Vec2Approx[T tNumber](a, b Vec2T[T]) bool {
	return Float(Abs(a.X()-b.X())) < FloatSmallestNonzero &&
		Float(Abs(a.Y()-b.Y())) < FloatSmallestNonzero
}

func Vec2ApproxTo[T tNumber](a, b Vec2T[T], delta T) bool {
	return Abs(a.X()-b.X()) < delta && Abs(a.Y()-b.Y()) < delta
}

func Vec2Min(a, b Vec2) Vec2 {
	return Vec2{
		min(a[Vx], b[Vx]),
		min(a[Vy], b[Vy]),
	}
}

func Vec2MinAbs(a, b Vec2) Vec2 {
	return Vec2{
		min(Abs(a[Vx]), Abs(b[Vx])),
		min(Abs(a[Vy]), Abs(b[Vy])),
	}
}

func Vec2Max(a, b Vec2) Vec2 {
	return Vec2{
		max(a[Vx], b[Vx]),
		max(a[Vy], b[Vy]),
	}
}

func Vec2MaxAbs[T tNumber](a, b Vec2T[T]) Vec2T[T] {
	return Vec2T[T]{
		max(T(Abs(a[Vx])), T(Abs(b[Vx]))),
		max(T(Abs(a[Vy])), T(Abs(b[Vy]))),
	}
}

func Vec2Dot[T tNumber](v, other Vec2T[T]) T {
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

func (v Vec2T[T]) X() T           { return v[Vx] }
func (v Vec2T[T]) Y() T           { return v[Vy] }
func (v Vec2T[T]) Width() T       { return v[Vx] }
func (v Vec2T[T]) Height() T      { return v[Vy] }
func (v *Vec2T[T]) PX() *T        { return &v[Vx] }
func (v *Vec2T[T]) PY() *T        { return &v[Vy] }
func (v *Vec2T[T]) SetX(x T)      { v[Vx] = x }
func (v *Vec2T[T]) SetY(y T)      { v[Vy] = y }
func (v *Vec2T[T]) SetWidth(x T)  { v[Vx] = x }
func (v *Vec2T[T]) SetHeight(y T) { v[Vy] = y }
func (v Vec2T[T]) AsVec3() Vec3   { return NewVec3(v[Vx], v[Vy], 0) }
func (v Vec2T[T]) XY() (T, T)     { return v[Vx], v[Vy] }

func (v Vec2T[T]) AsVec2i() Vec2i {
	return Vec2i{int32(v[Vx]), int32(v[Vy])}
}

func (v Vec2T[T]) Add(other Vec2T[T]) Vec2T[T] {
	return Vec2T[T]{v[Vx] + other[Vx], v[Vy] + other[Vy]}
}

func (v *Vec2T[T]) AddAssign(other Vec2T[T]) {
	v[Vx] += other[Vx]
	v[Vy] += other[Vy]
}

func (v Vec2T[T]) Subtract(other Vec2T[T]) Vec2T[T] {
	return Vec2T[T]{v[Vx] - other[Vx], v[Vy] - other[Vy]}
}

func (v *Vec2T[T]) SubtractAssign(other Vec2T[T]) {
	v[Vx] -= other[Vx]
	v[Vy] -= other[Vy]
}

func (v Vec2T[T]) Multiply(other Vec2T[T]) Vec2T[T] {
	return Vec2T[T]{v[Vx] * other[Vx], v[Vy] * other[Vy]}
}

func (v *Vec2T[T]) MultiplyAssign(other Vec2T[T]) {
	v[Vx] *= other[Vx]
	v[Vy] *= other[Vy]
}

func (v Vec2T[T]) Divide(other Vec2T[T]) Vec2T[T] {
	return Vec2T[T]{v[Vx] / other[Vx], v[Vy] / other[Vy]}
}

func (v *Vec2T[T]) DivideAssign(other Vec2T[T]) {
	v[Vx] /= other[Vx]
	v[Vy] /= other[Vy]
}

func (v Vec2T[T]) Scale(scalar T) Vec2T[T] {
	return Vec2T[T]{v[Vx] * scalar, v[Vy] * scalar}
}

func (v *Vec2T[T]) ScaleAssign(scalar T) {
	v[Vx] *= scalar
	v[Vy] *= scalar
}

func (v Vec2T[T]) Shrink(scalar T) Vec2T[T] {
	return Vec2T[T]{v[Vx] / scalar, v[Vy] / scalar}
}

func (v *Vec2T[T]) ShrinkAssign(scalar T) {
	v[Vx] /= scalar
	v[Vy] /= scalar
}

func (v Vec2T[T]) Length() T {
	return T(Sqrt(Vec2Dot(v, v)))
}

func (v Vec2T[T]) Normal() Vec2T[T] {
	return v.Scale(1.0 / v.Length())
}

func (v *Vec2T[T]) Normalize() {
	v.ScaleAssign(1.0 / v.Length())
}

func (v Vec2T[T]) Negative() Vec2T[T] {
	return Vec2T[T]{-v[Vx], -v[Vy]}
}

func (v *Vec2T[T]) Inverse() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
}

func (v Vec2T[T]) Abs() Vec2T[T] {
	return Vec2T[T]{T(Abs(v[Vx])), T(Abs(v[Vy]))}
}

func (v Vec2T[T]) Distance(other Vec2T[T]) T {
	return v.Subtract(other).Length()
}

func (v Vec2T[T]) String() string {
	return fmt.Sprintf(vec2StrFmt, v[Vx], v[Vy])
}

func (v Vec2T[T]) Angle(other Vec2T[T]) T {
	return Acos(Vec2Dot(v, other) / (v.Length() * other.Length()))
}

func (v Vec2T[T]) Equals(other Vec2T[T]) bool {
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

func (v Vec2T[T]) LargestAxis() T {
	return max(v[Vx], v[Vy])
}

func (v Vec2T[T]) LargestAxisDelta() T {
	lo := min(v[Vx], v[Vy])
	hi := max(v[Vx], v[Vy])
	if Abs(lo) > Abs(hi) {
		return lo
	} else {
		return hi
	}
}

func Vec2Inf(sign int) Vec2 {
	return Vec2{Inf(sign), Inf(sign)}
}

func Vec2NaN() Vec2 {
	return Vec2{NaN(), NaN()}
}

func (v Vec2T[T]) IsZero() bool {
	return Vec2Approx(v, Vec2T[T]{0, 0})
}

func (v Vec2T[T]) IsInf(sign int) bool {
	return IsInf(v[Vx], sign) || IsInf(v[Vy], sign)
}

func (v Vec2T[T]) IsNaN() bool {
	return IsNaN(v[Vx]) || IsNaN(v[Vy])
}

func (v Vec2T[T]) IsValidNonZero() bool {
	return !v.IsZero() && !v.IsNaN() && !v.IsInf(-1) && !v.IsInf(1)
}
