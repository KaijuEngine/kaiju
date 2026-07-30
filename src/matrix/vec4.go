/******************************************************************************/
/* vec4.go                                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"fmt"
)

const vec4StrFmt = "%f, %f, %f, %f"

type Vec4T[T tNumber] [4]T
type Vec4 = Vec4T[Float]

func (v Vec4T[T]) X() T               { return v[Vx] }
func (v Vec4T[T]) Y() T               { return v[Vy] }
func (v Vec4T[T]) Z() T               { return v[Vz] }
func (v Vec4T[T]) W() T               { return v[Vw] }
func (v Vec4T[T]) Left() T            { return v[Vx] }
func (v Vec4T[T]) Top() T             { return v[Vy] }
func (v Vec4T[T]) Right() T           { return v[Vz] }
func (v Vec4T[T]) Bottom() T          { return v[Vw] }
func (v Vec4T[T]) Width() T           { return v[Vz] }
func (v Vec4T[T]) Height() T          { return v[Vw] }
func (v *Vec4T[T]) PX() *T            { return &v[Vx] }
func (v *Vec4T[T]) PY() *T            { return &v[Vy] }
func (v *Vec4T[T]) PZ() *T            { return &v[Vz] }
func (v *Vec4T[T]) PW() *T            { return &v[Vw] }
func (v *Vec4T[T]) SetX(x T)          { v[Vx] = x }
func (v *Vec4T[T]) SetY(y T)          { v[Vy] = y }
func (v *Vec4T[T]) SetZ(z T)          { v[Vz] = z }
func (v *Vec4T[T]) SetW(w T)          { v[Vw] = w }
func (v *Vec4T[T]) SetLeft(x T)       { v[Vx] = x }
func (v *Vec4T[T]) SetTop(y T)        { v[Vy] = y }
func (v *Vec4T[T]) SetRight(z T)      { v[Vz] = z }
func (v *Vec4T[T]) SetBottom(w T)     { v[Vw] = w }
func (v *Vec4T[T]) SetWidth(z T)      { v[Vz] = z }
func (v *Vec4T[T]) SetHeight(w T)     { v[Vw] = w }
func (v Vec4T[T]) AsVec3() Vec3       { return Vec3{Float(v[Vx]), Float(v[Vy]), Float(v[Vz])} }
func (v Vec4T[T]) XYZW() (T, T, T, T) { return v[Vx], v[Vy], v[Vz], v[Vw] }
func (v Vec4T[T]) Horizontal() T      { return v[Vx] + v[Vz] }
func (v Vec4T[T]) Vertical() T        { return v[Vy] + v[Vw] }

func (v Vec4T[T]) AsVec4i() Vec4i {
	return Vec4i{int32(v[Vx]), int32(v[Vy]), int32(v[Vz]), int32(v[Vw])}
}

func NewVec4T[T1, T2, T3, T4, T5 tNumber](x T2, y T3, z T4, w T5) Vec4T[T1] {
	return Vec4T[T1]{T1(x), T1(y), T1(z), T1(w)}
}

func NewVec4[T1, T2, T3, T4 tNumber](x T1, y T2, z T3, w T4) Vec4 {
	return NewVec4T[Float](x, y, z, w)
}

func Vec4FromArray[T tNumber](a [4]T) Vec4 {
	return Vec4{Float(a[0]), Float(a[1]), Float(a[2]), Float(a[3])}
}

func Vec4FromSlice[T tNumber](a []T) Vec4 {
	return Vec4{Float(a[0]), Float(a[1]), Float(a[2]), Float(a[3])}
}

func (v Vec4T[T]) Add(other Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{v[Vx] + other[Vx], v[Vy] + other[Vy], v[Vz] + other[Vz], v[Vw] + other[Vw]}
}

func (v *Vec4T[T]) AddAssign(other Vec4T[T]) {
	v[Vx] += other[Vx]
	v[Vy] += other[Vy]
	v[Vz] += other[Vz]
	v[Vw] += other[Vw]
}

func (v Vec4T[T]) Subtract(other Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{v[Vx] - other[Vx], v[Vy] - other[Vy], v[Vz] - other[Vz], v[Vw] - other[Vw]}
}

func (v *Vec4T[T]) SubtractAssign(other Vec4T[T]) {
	v[Vx] -= other[Vx]
	v[Vy] -= other[Vy]
	v[Vz] -= other[Vz]
	v[Vw] -= other[Vw]
}

func (v Vec4T[T]) Multiply(other Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{v[Vx] * other[Vx], v[Vy] * other[Vy], v[Vz] * other[Vz], v[Vw] * other[Vw]}
}

func (v *Vec4T[T]) MultiplyAssign(other Vec4T[T]) {
	v[Vx] *= other[Vx]
	v[Vy] *= other[Vy]
	v[Vz] *= other[Vz]
	v[Vw] *= other[Vw]
}

func (v Vec4T[T]) Divide(other Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{v[Vx] / other[Vx], v[Vy] / other[Vy], v[Vz] / other[Vz], v[Vw] / other[Vw]}
}

func (v *Vec4T[T]) DivideAssign(other Vec4T[T]) {
	v[Vx] /= other[Vx]
	v[Vy] /= other[Vy]
	v[Vz] /= other[Vz]
	v[Vw] /= other[Vw]
}

func (v Vec4T[T]) Scale(scalar T) Vec4T[T] {
	return Vec4T[T]{v[Vx] * scalar, v[Vy] * scalar, v[Vz] * scalar, v[Vw] * scalar}
}

func (v *Vec4T[T]) ScaleAssign(scalar T) {
	v[Vx] *= scalar
	v[Vy] *= scalar
	v[Vz] *= scalar
	v[Vw] *= scalar
}

func (v Vec4T[T]) Shrink(scalar T) Vec4T[T] {
	return Vec4T[T]{v[Vx] / scalar, v[Vy] / scalar, v[Vz] / scalar, v[Vw] / scalar}
}

func (v *Vec4T[T]) ShrinkAssign(scalar T) {
	v[Vx] /= scalar
	v[Vy] /= scalar
	v[Vz] /= scalar
	v[Vw] /= scalar
}

func (v Vec4T[T]) Length() T {
	return T(Sqrt(Vec4Dot(v, v)))
}

func (v Vec4T[T]) Normal() Vec4T[T] {
	return v.Scale(1.0 / v.Length())
}

func (v *Vec4T[T]) Normalize() {
	v.ScaleAssign(1.0 / v.Length())
}

func (v Vec4T[T]) Negative() Vec4T[T] {
	return Vec4T[T]{-v[Vx], -v[Vy], -v[Vz], -v[Vw]}
}

func (v *Vec4T[T]) Inverse() {
	v[Vx] = -v[Vx]
	v[Vy] = -v[Vy]
	v[Vz] = -v[Vz]
	v[Vw] = -v[Vw]
}

func Vec4Roughly(a, b Vec4) bool {
	return Abs(a.X()-b.X()) < Roughly &&
		Abs(a.Y()-b.Y()) < Roughly &&
		Abs(a.Z()-b.Z()) < Roughly &&
		Abs(a.W()-b.W()) < Roughly
}

func Vec4Approx[T tNumber](a, b Vec4T[T]) bool {
	return Float(Abs(a.X()-b.X())) < FloatSmallestNonzero &&
		Float(Abs(a.Y()-b.Y())) < FloatSmallestNonzero &&
		Float(Abs(a.Z()-b.Z())) < FloatSmallestNonzero &&
		Float(Abs(a.W()-b.W())) < FloatSmallestNonzero
}

func Vec4ApproxTo[T tNumber](a, b Vec4T[T], delta T) bool {
	return Abs(a.X()-b.X()) < delta && Abs(a.Y()-b.Y()) < delta &&
		Abs(a.Z()-b.Z()) < delta && Abs(a.W()-b.W()) < delta
}

func Vec4Min[T tNumber](a, b Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{
		min(a[Vx], b[Vx]),
		min(a[Vy], b[Vy]),
		min(a[Vz], b[Vz]),
		min(a[Vw], b[Vw]),
	}
}

func Vec4MinAbs[T tNumber](a, b Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{
		min(Abs(a[Vx]), Abs(b[Vx])),
		min(Abs(a[Vy]), Abs(b[Vy])),
		min(Abs(a[Vz]), Abs(b[Vz])),
		min(Abs(a[Vw]), Abs(b[Vw])),
	}
}

func Vec4Max[T tNumber](a, b Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{
		max(a[Vx], b[Vx]),
		max(a[Vy], b[Vy]),
		max(a[Vz], b[Vz]),
		max(a[Vw], b[Vw]),
	}
}

func Vec4MaxAbs[T tNumber](a, b Vec4T[T]) Vec4T[T] {
	return Vec4T[T]{
		max(T(Abs(a[Vx])), T(Abs(b[Vx]))),
		max(T(Abs(a[Vy])), T(Abs(b[Vy]))),
		max(T(Abs(a[Vz])), T(Abs(b[Vz]))),
		max(T(Abs(a[Vw])), T(Abs(b[Vw]))),
	}
}

func (v Vec4T[T]) Abs() Vec4T[T] {
	return Vec4T[T]{T(Abs(v[Vx])), T(Abs(v[Vy])), T(Abs(v[Vz])), T(Abs(v[Vw]))}
}

func (v Vec4T[T]) Distance(other Vec4T[T]) T {
	return v.Subtract(other).Length()
}

func Vec4Dot[T tNumber](v, other Vec4T[T]) T {
	return v[Vx]*other[Vx] + v[Vy]*other[Vy] + v[Vz]*other[Vz] + v[Vw]*other[Vw]
}

func Vec4Lerp[T tNumber](from, to Vec4T[T], t T) Vec4T[T] {
	return from.Add(to.Subtract(from).Scale(t))
}

func Vec4FromString(str string) Vec4 {
	var v Vec4
	fmt.Sscanf(str, vec4StrFmt, &v[Vx], &v[Vy], &v[Vz], &v[Vw])
	return v
}

func (v Vec4T[T]) String() string {
	return fmt.Sprintf(vec4StrFmt, Float(v[Vx]), Float(v[Vy]), Float(v[Vz]), Float(v[Vw]))
}

func (v Vec4T[T]) Angle(other Vec4T[T]) T {
	return Acos(Vec4Dot(v, other) / (v.Length() * other.Length()))
}

func (v Vec4T[T]) Equals(other Vec4T[T]) bool {
	return Vec4Approx(v, other)
}

func Vec4Zero() Vec4 { return Vec4{0, 0, 0, 0} }
func Vec4One() Vec4  { return Vec4{1, 1, 1, 1} }
func Vec4Half() Vec4 { return Vec4{0.5, 0.5, 0.5, 0.5} }
func Vec4Largest() Vec4 {
	return Vec4{FloatMax, FloatMax, FloatMax, FloatMax}
}

func (v Vec4T[T]) LargestAxis() T {
	return max(v[Vx], v[Vy], v[Vz], v[Vw])
}

func (v Vec4T[T]) LargestAxisDelta() T {
	lo := min(v[Vx], v[Vy], v[Vz], v[Vw])
	hi := max(v[Vx], v[Vy], v[Vz], v[Vw])
	if Abs(lo) > Abs(hi) {
		return lo
	} else {
		return hi
	}
}

func Vec4Area[T tNumber](xa, ya, xb, yb T) Vec4 {
	return NewVec4(min(xa, xb), min(ya, yb), max(xa, xb), max(ya, yb))
}

func (v Vec4T[T]) BoxContains(x, y Float) bool {
	return Float(v.X()) <= x && Float(v.X())+Float(v.Width()) >= x && Float(v.Y()) <= y && Float(v.Y())+Float(v.Height()) >= y
}

func (v Vec4T[T]) AreaContains(x, y Float) bool {
	return Float(v.X()) <= x && Float(v.Right()) >= x && Float(v.Y()) <= y && Float(v.Bottom()) >= y
}

func (v Vec4T[T]) ScreenAreaContains(x, y Float) bool {
	return Float(v.X()) <= x && Float(v.Right()) >= x && Float(v.Y()) <= y && Float(v.Bottom()) >= y
}
