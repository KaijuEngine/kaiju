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

type FloatingPoint interface {
	~float32 | ~float64
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface {
	Signed | Unsigned
}

type Number interface {
	Integer | FloatingPoint
}

type Vector interface {
	Vec2 | Vec3 | Vec4 | Quaternion
}

type Matrix interface {
	Mat3 | Mat4
}

func Rad2Deg[T FloatingPoint](radian T) T {
	return radian * (180.0 / math.Pi)
}

func Deg2Rad[T FloatingPoint](degree T) T {
	return degree * (math.Pi / 180.0)
}

func Clamp[T FloatingPoint](current, minimum, maximum T) T {
	return T(max(minimum, min(maximum, current)))
}
