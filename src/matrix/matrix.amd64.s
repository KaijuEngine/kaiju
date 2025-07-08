/******************************************************************************/
/* mat4.amd64.s                                                               */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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

#define DOT(a, b, to) \
	MOVUPS a,  X0     \
	MULPS  b,  X0     \
	HADDPS X0, X0     \
	HADDPS X0, X0     \
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
	PACK_COLUMNS(64)
	DOT(a+0(FP), X1, ret+128(FP))   // x0y0
	DOT(a+0(FP), X2, ret+132(FP))   // x0y1
	DOT(a+0(FP), X3, ret+136(FP))   // x0y2
	DOT(a+0(FP), X4, ret+140(FP))   // x0y3
	DOT(a+16(FP), X1, ret+144(FP))  // x1y0
	DOT(a+16(FP), X2, ret+148(FP))  // x1y1
	DOT(a+16(FP), X3, ret+152(FP))  // x1y2
	DOT(a+16(FP), X4, ret+156(FP))  // x1y3
	DOT(a+32(FP), X1, ret+160(FP))  // x2y0
	DOT(a+32(FP), X2, ret+164(FP))  // x2y1
	DOT(a+32(FP), X3, ret+168(FP))  // x2y2
	DOT(a+32(FP), X4, ret+172(FP))  // x2y3
	DOT(a+48(FP), X1, ret+176(FP))  // x3y0
	DOT(a+48(FP), X2, ret+180(FP))  // x3y1
	DOT(a+48(FP), X3, ret+184(FP))  // x3y2
	DOT(a+48(FP), X4, ret+188(FP))  // x3y3
	RET

// func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4
TEXT   ·Mat4MultiplyVec4(SB),NOSPLIT,$0-96
	PACK_COLUMNS(0)
	DOT(b+64(FP), X1, ret+80(FP))   // x
	DOT(b+64(FP), X2, ret+84(FP))   // y
	DOT(b+64(FP), X3, ret+88(FP))   // z
	DOT(b+64(FP), X4, ret+92(FP))   // w
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
