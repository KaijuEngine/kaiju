package matrix

import (
	"math"
)

type VectorComponent = int
type QuaternionComponent = int

const (
	Vx VectorComponent = iota
	Vy
	Vz
	Vw
)

const (
	Qw QuaternionComponent = iota
	Qx
	Qy
	Qz
)

type tFloatingPoint interface {
	~float32 | ~float64
}

type tSigned interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type tUnsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type tInteger interface {
	tSigned | tUnsigned
}

type tNumber interface {
	tInteger | tFloatingPoint
}

type tVector interface {
	Vec2 | Vec3 | Vec4 | Quaternion
}

type tMatrix interface {
	Mat3 | Mat4
}

func Rad2Deg(radian Float) Float {
	return radian * (180.0 / math.Pi)
}

func Deg2Rad(degree Float) Float {
	return degree * (math.Pi / 180.0)
}

func Approx(a, b Float) bool {
	return math.Abs(float64(a-b)) < FloatSmallestNonzero
}

func clamp[T tFloatingPoint](current, minimum, maximum T) T {
	return T(max(minimum, min(maximum, current)))
}

func AbsInt(a int) int { return a & int(^uint(0)>>1) }
