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

func Vec3Min(a, b Vec3) Vec3 {
	return Vec3{
		Min(a[Vx], b[Vx]),
		Min(a[Vy], b[Vy]),
		Min(a[Vz], b[Vz]),
	}
}

func Vec3MinAbs(a, b Vec3) Vec3 {
	return Vec3{
		Min(Abs(a[Vx]), Abs(b[Vx])),
		Min(Abs(a[Vy]), Abs(b[Vy])),
		Min(Abs(a[Vz]), Abs(b[Vz])),
	}
}

func Vec3Max(a, b Vec3) Vec3 {
	return Vec3{
		Max(a[Vx], b[Vx]),
		Max(a[Vy], b[Vy]),
		Max(a[Vz], b[Vz]),
	}
}

func Vec3MaxAbs(a, b Vec3) Vec3 {
	return Vec3{
		Max(Abs(a[Vx]), Abs(b[Vx])),
		Max(Abs(a[Vy]), Abs(b[Vy])),
		Max(Abs(a[Vz]), Abs(b[Vz])),
	}
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
func Vec3Largest() Vec3  { return Vec3{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32} }

func (v Vec3) SquareDistance(b Vec3) Float {
	return (v[Vx]-b[Vx])*(v[Vx]-b[Vx]) + (v[Vy]-b[Vy])*(v[Vy]-b[Vy]) + (v[Vz]-b[Vz])*(v[Vz]-b[Vz])
}

func (v Vec3) LargestAxis() Float {
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
