/******************************************************************************/
/* vec4.amd64.s                                                               */
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

// func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4
TEXT   ·Vec4MultiplyMat4(SB),NOSPLIT,$0-96
	MOVUPS    a+0(FP),  X0
// x
	MOVUPS    b+16(FP), X1
	MULPS     X1,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+80(FP)
// y
	MOVUPS    a+0(FP),  X0
	MOVUPS    b+32(FP), X1
	MULPS     X1,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+84(FP)
// z
	MOVUPS    a+0(FP),  X0
	MOVUPS    b+48(FP), X1
	MULPS     X1,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+88(FP)
// w
	MOVUPS    a+0(FP),  X0
	MOVUPS    b+64(FP), X1
	MULPS     X1,      X0
	HADDPS    X0,      X0
	HADDPS    X0,      X0
	MOVUPS    X0,      ret+92(FP)
	RET
