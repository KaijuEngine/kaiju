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

func Vec2Approx(a, b Vec2) bool {
	return Abs(a.X()-b.X()) < math.SmallestNonzeroFloat32 &&
		Abs(a.Y()-b.Y()) < math.SmallestNonzeroFloat32
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
func Vec2Largest() Vec2 { return Vec2{math.MaxFloat32, math.MaxFloat32} }

func (v Vec2) LargestAxis() Float {
	return max(v[Vx], v[Vy])
}
