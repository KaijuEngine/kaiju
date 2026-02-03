/******************************************************************************/
/* matrix.amd64.s                                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

//go:build amd64

#include "textflag.h"

// func Mat4Multiply(a, b Mat4) Mat4
TEXT   ·Mat4Multiply(SB),NOSPLIT,$0-192
	// Load b rows (contiguous)
	MOVUPS b+64(FP), X1   // b row0
	MOVUPS b+80(FP), X2   // b row1
	MOVUPS b+96(FP), X3   // b row2
	MOVUPS b+112(FP), X4  // b row3
	// Compute ret row0 = sum (a row0[k] * b row k for k=0..3)
	MOVUPS a+0(FP), X0    // a row0: a00 a01 a02 a03
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5  // a00 a00 a00 a00
	MULPS  X1, X5         // a00 * b row0
	MOVAPS X0, X6
	SHUFPS $0x55, X6, X6  // a01 a01 a01 a01
	MULPS  X2, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xAA, X6, X6  // a02 a02 a02 a02
	MULPS  X3, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xFF, X6, X6  // a03 a03 a03 a03
	MULPS  X4, X6
	ADDPS  X6, X5
	MOVUPS X5, ret+128(FP)
	// Compute ret row1
	MOVUPS a+16(FP), X0   // a row1: a10 a11 a12 a13
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5
	MULPS  X1, X5
	MOVAPS X0, X6
	SHUFPS $0x55, X6, X6
	MULPS  X2, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xAA, X6, X6
	MULPS  X3, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xFF, X6, X6
	MULPS  X4, X6
	ADDPS  X6, X5
	MOVUPS X5, ret+144(FP)
	// Compute ret row2
	MOVUPS a+32(FP), X0   // a row2: a20 a21 a22 a23
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5
	MULPS  X1, X5
	MOVAPS X0, X6
	SHUFPS $0x55, X6, X6
	MULPS  X2, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xAA, X6, X6
	MULPS  X3, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xFF, X6, X6
	MULPS  X4, X6
	ADDPS  X6, X5
	MOVUPS X5, ret+160(FP)
	// Compute ret row3
	MOVUPS a+48(FP), X0   // a row3: a30 a31 a32 a33
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5
	MULPS  X1, X5
	MOVAPS X0, X6
	SHUFPS $0x55, X6, X6
	MULPS  X2, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xAA, X6, X6
	MULPS  X3, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xFF, X6, X6
	MULPS  X4, X6
	ADDPS  X6, X5
	MOVUPS X5, ret+176(FP)
	RET

// func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4
TEXT   ·Mat4MultiplyVec4(SB),NOSPLIT,$0-96
	// Load b rows
	MOVUPS b+16(FP), X1  // b row0
	MOVUPS b+32(FP), X2  // b row1
	MOVUPS b+48(FP), X3  // b row2
	MOVUPS b+64(FP), X4  // b row3
	// Load a vec
	MOVUPS a+0(FP), X0   // ax ay az aw
	// Compute ret = ax * row0 + ay * row1 + az * row2 + aw * row3
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5  // ax ax ax ax
	MULPS  X1, X5         // ax * b row0
	MOVAPS X0, X6
	SHUFPS $0x55, X6, X6  // ay ay ay ay
	MULPS  X2, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xAA, X6, X6  // az az az az
	MULPS  X3, X6
	ADDPS  X6, X5
	MOVAPS X0, X6
	SHUFPS $0xFF, X6, X6  // aw aw aw aw
	MULPS  X4, X6
	ADDPS  X6, X5
	MOVUPS X5, ret+80(FP)
	RET

// func Vec4MultiplyMat4(a Vec4, b Mat4) Vec4
TEXT   ·Vec4MultiplyMat4(SB),NOSPLIT,$0-96
	// Load a rows
	MOVUPS a+0(FP), X1   // row0
	MOVUPS a+16(FP), X2  // row1
	MOVUPS a+32(FP), X3  // row2
	MOVUPS a+48(FP), X4  // row3
	// Transpose to get "columns" (m00 m10 m20 m30, etc.)
	MOVAPS X1, X5
	UNPCKLPS X2, X5      // m00 m10 m01 m11
	UNPCKHPS X1, X2      // m02 m12 m03 m13
	MOVAPS X3, X6
	UNPCKLPS X4, X6      // m20 m30 m21 m31
	UNPCKHPS X3, X4      // m22 m32 m23 m33
	MOVAPS X5, X7
	UNPCKLPS X6, X7      // m00 m10 m20 m30  (col0)
	UNPCKHPS X5, X6      // m01 m11 m21 m31  (col1)
	MOVAPS X2, X8
	UNPCKLPS X4, X8      // m02 m12 m22 m32  (col2)
	UNPCKHPS X2, X4      // m03 m13 m23 m33  (col3)
	// Load b vec
	MOVUPS b+64(FP), X0  // bx by bz bw
	// Compute ret = bx * col0 + by * col1 + bz * col2 + bw * col3
	MOVAPS X0, X5
	SHUFPS $0x00, X5, X5  // bx bx bx bx
	MULPS  X7, X5         // bx * col0
	MOVAPS X0, X9
	SHUFPS $0x55, X9, X9  // by by by by
	MULPS  X6, X9
	ADDPS  X9, X5
	MOVAPS X0, X9
	SHUFPS $0xAA, X9, X9  // bz bz bz bz
	MULPS  X8, X9
	ADDPS  X9, X5
	MOVAPS X0, X9
	SHUFPS $0xFF, X9, X9  // bw bw bw bw
	MULPS  X4, X9
	ADDPS  X9, X5
	MOVUPS X5, ret+80(FP)
	RET
