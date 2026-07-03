/******************************************************************************/
/* matrix_config.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import (
	"math"
)

type ColorComponent = int
type VectorComponent = int
type QuaternionComponent = int

const (
	R ColorComponent = iota
	G
	B
	A
)

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

const RadToDegVal = (180.0 / math.Pi)
const DegToRadVal = (math.Pi / 180.0)

func Rad2Deg[T tNumber](radian T) Float {
	return Float(radian) * Float(180.0/math.Pi)
}

func Deg2Rad[T tNumber](degree T) Float {
	return Float(degree) * Float(math.Pi/180.0)
}

func Approx[T tNumber](a, b T) bool {
	return math.Abs(float64(a)-float64(b)) < float64(FloatSmallestNonzero)
}

func ApproxTo[T tNumber](a, b, tolerance T) bool {
	return math.Abs(float64(a)-float64(b)) <= float64(tolerance)
}

func Clamp[T tNumber](current, minimum, maximum T) Float {
	return max(Float(minimum), min(Float(maximum), Float(current)))
}

func AbsInt(a int) int { return a & int(^uint(0)>>1) }
