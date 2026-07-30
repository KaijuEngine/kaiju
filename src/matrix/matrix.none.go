//go:build !amd64 && !arm64

/******************************************************************************/
/* matrix.none.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

func Mat4fMultiply(a, b Mat4f) Mat4f {
	return mat4fMultiplyFallback(a, b)
}

func Mat4fMultiplyVec4f(a Mat4f, b Vec4f) Vec4f {
	return mat4fMultiplyVec4fFallback(a, b)
}

func Vec4fMultiplyMat4f(v Vec4f, m Mat4f) Vec4f {
	return vec4fMultiplyMat4fFallback(v, m)
}
