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

#define DOT(a, b, to)    \
	MOVUPS a,  X0        \
	MULPS  b,  X0        \
	MOVAPS X0, X5        \
	SHUFPS $0x4E, X5, X5 \
	ADDPS  X5, X0        \
	MOVAPS X0, X5        \
	SHUFPS $0xB1, X5, X5 \
	ADDPS  X5, X0        \
	MOVSS  X0, to

#define PACK_COLUMNS(start)           \
	INSERTPS  $14, m+start+0(FP),  X1 \  // x0y0
	INSERTPS  $14, m+start+4(FP),  X2 \  // x0y1
	INSERTPS  $14, m+start+8(FP),  X3 \  // x0y2
	INSERTPS  $14, m+start+12(FP), X4 \  // x0y3
	INSERTPS  $16, m+start+16(FP), X1 \  // x1y0
	INSERTPS  $16, m+start+20(FP), X2 \  // x1y1
	INSERTPS  $16, m+start+24(FP), X3 \  // x1y2
	INSERTPS  $16, m+start+28(FP), X4 \  // x1y3
	INSERTPS  $32, m+start+32(FP), X1 \  // x2y0
	INSERTPS  $32, m+start+36(FP), X2 \  // x2y1
	INSERTPS  $32, m+start+40(FP), X3 \  // x2y2
	INSERTPS  $32, m+start+44(FP), X4 \  // x2y3
	INSERTPS  $48, m+start+48(FP), X1 \  // x3y0
	INSERTPS  $48, m+start+52(FP), X2 \  // x3y1
	INSERTPS  $48, m+start+56(FP), X3 \  // x3y2
	INSERTPS  $48, m+start+60(FP), X4    // x3y3

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
	MOVUPS m+0(FP), X0
	MOVUPS m+16(FP), X1
	MOVUPS m+32(FP), X2
	MOVUPS m+48(FP), X3
	MOVAPS X0, X4
	UNPCKLPS X1, X4
	MOVAPS X0, X5
	UNPCKHPS X1, X5
	MOVAPS X2, X6
	UNPCKLPS X3, X6
	MOVAPS X2, X7
	UNPCKHPS X3, X7
	MOVAPS X4, X8
	MOVLHPS X6, X8
	MOVAPS X4, X9
	MOVHLPS X4, X9
	MOVAPS X6, X12
	MOVHLPS X6, X12
	MOVLHPS X12, X9
	MOVAPS X5, X10
	MOVLHPS X7, X10
	MOVAPS X5, X11
	MOVHLPS X5, X11
	MOVAPS X7, X13
	MOVHLPS X7, X13
	MOVLHPS X13, X11
	DOT(b+64(FP), X8, ret+80(FP))   // x
	DOT(b+64(FP), X9, ret+84(FP))   // y
	DOT(b+64(FP), X10, ret+88(FP))  // z
	DOT(b+64(FP), X11, ret+92(FP))  // w
	RET

// func Vec4MultiplyMat4(a Vec4, b Mat4) Vec4
TEXT   ·Vec4MultiplyMat4(SB),NOSPLIT,$0-96
	MOVUPS    b+16(FP), X1
	MOVUPS    b+32(FP), X2
	MOVUPS    b+48(FP), X3
	MOVUPS    b+64(FP), X4
	DOT(a+0(FP), X1, ret+80(FP))    // x
	DOT(a+0(FP), X2, ret+84(FP))    // y
	DOT(a+0(FP), X3, ret+88(FP))    // z
	DOT(a+0(FP), X4, ret+92(FP))    // w
	RET
