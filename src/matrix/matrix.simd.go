//go:build amd64 || arm64

/******************************************************************************/
/* matrix.simd.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

//go:noescape
func Mat4Multiply(a, b Mat4) Mat4

//go:noescape
func Mat4MultiplyAVX(a, b Mat4) Mat4

//go:noescape
func Mat4MultiplyAVX512(a, b Mat4) Mat4

//go:noescape
func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4

//go:noescape
func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4
