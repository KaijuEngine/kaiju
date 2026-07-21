//go:build !F64

/******************************************************************************/
/* float32.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import "math"

type Float = float32
type Mat4f = Mat4
type Mat3f = Mat3
type Vec4f = Vec4
type Vec3f = Vec3

const FloatSmallestNonzero = Float(math.SmallestNonzeroFloat32)
const FloatMax = Float(math.MaxFloat32)

func Abs[T tNumber](x T) T {
	return T(math.Float32frombits(math.Float32bits(float32(x)) &^ (1 << 31)))
}

func Mat4Multiply(a, b Mat4f) Mat4f {
	return Mat4fMultiply(a, b)
}

func Mat4MultiplyVec4(a Mat4f, b Vec4f) Vec4f {
	return Mat4fMultiplyVec4f(a, b)
}

func Vec4MultiplyMat4(v Vec4f, m Mat4f) Vec4f {
	return Vec4fMultiplyMat4f(v, m)
}
