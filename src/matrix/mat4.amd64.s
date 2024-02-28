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
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

//go:build amd64

#include "textflag.h"

// func Mat4Multiply(a, b Mat4) Mat4
TEXT   ·Mat4Multiply(SB),NOSPLIT,$0-192
	INSERTPS  $14, b+64(FP),  X1	// x0y0
	INSERTPS  $14, b+68(FP),  X2	// x0y1
	INSERTPS  $14, b+72(FP),  X3	// x0y2
	INSERTPS  $14, b+76(FP),  X4	// x0y3
	INSERTPS  $16, b+80(FP),  X1	// x1y0	
	INSERTPS  $16, b+84(FP),  X2	// x1y1
	INSERTPS  $16, b+88(FP),  X3	// x1y2
	INSERTPS  $16, b+92(FP),  X4	// x1y3
	INSERTPS  $32, b+96(FP),  X1	// x2y0
	INSERTPS  $32, b+100(FP), X2	// x2y1
	INSERTPS  $32, b+104(FP), X3	// x2y2
	INSERTPS  $32, b+108(FP), X4	// x2y3
	INSERTPS  $48, b+112(FP), X1	// x3y0
	INSERTPS  $48, b+116(FP), X2	// x3y1
	INSERTPS  $48, b+120(FP), X3	// x3y2
	INSERTPS  $48, b+124(FP), X4	// x3y3
// x0y0
	MOVUPS    a+0(FP), X0
	MULPS     X1,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+128(FP)
// x0y1
	MOVUPS    a+0(FP), X0
	MULPS     X2,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+132(FP)
// x0y2
	MOVUPS    a+0(FP), X0
	MULPS     X3,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+136(FP)
// x0y3
	MOVUPS    a+0(FP), X0
	MULPS     X4,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+140(FP)
// x1y0
	MOVUPS    a+16(FP), X0
	MULPS     X1,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+144(FP)
// x1y1
	MOVUPS    a+16(FP), X0
	MULPS     X2,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+148(FP)
// x1y2
	MOVUPS    a+16(FP), X0
	MULPS     X3,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+152(FP)
// x1y3
	MOVUPS    a+16(FP), X0
	MULPS     X4,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+156(FP)
// x2y0
	MOVUPS    a+32(FP), X0
	MULPS     X1,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+160(FP)
// x2y1
	MOVUPS    a+32(FP), X0
	MULPS     X2,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+164(FP)
// x2y2
	MOVUPS    a+32(FP), X0
	MULPS     X3,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+168(FP)
// x2y3
	MOVUPS    a+32(FP), X0
	MULPS     X4,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+172(FP)
// x3y0
	MOVUPS    a+48(FP), X0
	MULPS     X1,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+176(FP)
// x3y1
	MOVUPS    a+48(FP), X0
	MULPS     X2,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+180(FP)
// x3y2
	MOVUPS    a+48(FP), X0
	MULPS     X3,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+184(FP)
// x3y3
	MOVUPS    a+48(FP), X0
	MULPS     X4,       X0
	HADDPS    X0,       X0
	HADDPS    X0,       X0
	MOVUPS    X0,       ret+188(FP)
	RET
