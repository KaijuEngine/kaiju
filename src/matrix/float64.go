//go:build F64

/******************************************************************************/
/* float64.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import "math"

type Float = float64

const FloatSmallestNonzero = Float(math.SmallestNonzeroFloat64)
const FloatMax = Float(math.MaxFloat64)

func Abs[T tNumber](x T) T {
	return T(math.Abs(float64(x)))
}

func Mat4Multiply(a, b Mat4) Mat4 {
	return mat4MultiplyFallback(a, b)
}

func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4 {
	return mat4MultiplyVec4Fallback(a, b)
}

func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4 {
	return vec4MultiplyMat4Fallback(v, m)
}
