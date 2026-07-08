//go:build (amd64 || arm64) && F64

/******************************************************************************/
/* matrix.simd.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

//go:noescape
func Mat4fMultiply(a, b Mat4f) Mat4f

//go:noescape
func Mat4fMultiplyVec4f(a Mat4f, b Vec4f) Vec4f

//go:noescape
func Vec4fMultiplyMat4f(v Vec4f, m Mat4f) Vec4f
