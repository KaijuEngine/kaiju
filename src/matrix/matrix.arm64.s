/******************************************************************************/
/* matrix.arm64.s                                                             */
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

//go:build arm64

#include "textflag.h"

#define MULSUM4ROWS() \
    WORD $0x6e24dc08  \  // fmul.4s v8, v0, v4
    WORD $0x6e25dc29  \  // fmul.4s v9, v1, v5
    WORD $0x6e26dc4a  \  // fmul.4s v10, v2, v6
    WORD $0x6e27dc6b  \  // fmul.4s v11, v3, v7
    WORD $0x4e29d50c  \  // fadd.4s v12, v8, v9
    WORD $0x4e2ad56d  \  // fadd.4s v13, v11, v10
    WORD $0x4e2cd5ae     // fadd.4s v14, v13, v12

#define MULROWWISE(inOff, outOff) \
    FMOVQ a_rI+inOff(FP), F4      \
    VDUP V4.S[1], V5.S4           \
    VDUP V4.S[2], V6.S4           \
    VDUP V4.S[3], V7.S4           \
    VDUP V4.S[0], V4.S4           \
    MULSUM4ROWS()                 \
    FMOVQ F14, ret_rI+outOff(FP)        

// func Mat4Multiply(a, b Mat4) Mat4
TEXT   ·Mat4Multiply(SB),NOSPLIT,$0-192
    FMOVQ b_0+64(FP), F0
    FMOVQ b_4+80(FP), F1
    FMOVQ b_8+96(FP), F2
    FMOVQ b_12+112(FP), F3
    MULROWWISE(0, 128)
    MULROWWISE(16, 144)
    MULROWWISE(32, 160)
    MULROWWISE(48, 176)
    RET

// func Mat4MultiplyVec4(a Mat4, b Vec4) Vec4
TEXT   ·Mat4MultiplyVec4(SB),NOSPLIT,$0-96
    FMOVQ a_0+0(FP), F0
    FMOVQ a_4+16(FP), F1
    FMOVQ a_8+32(FP), F2
    FMOVQ a_12+48(FP), F3
    FMOVQ b+64(FP), F4
    VDUP V4.S[1], V5.S4
    VDUP V4.S[2], V6.S4
    VDUP V4.S[3], V7.S4
    VDUP V4.S[0], V4.S4
    MULSUM4ROWS()
    FMOVQ F14, ret+80(FP)
    RET

// func Vec4MultiplyMat4(v Vec4, m Mat4) Vec4
TEXT   ·Vec4MultiplyMat4(SB),NOSPLIT,$0-96
    FMOVQ m_0+16(FP), F0
    FMOVQ m_4+32(FP), F1
    FMOVQ m_8+48(FP), F2
    FMOVQ m_12+64(FP), F3
    FMOVQ v+0(FP), F4

    WORD $0x6e24dc05          // fmul.4s v5, v0, v4
    WORD $0x6e24dc26          // fmul.4s v6, v1, v4
    WORD $0x6e24dc47          // fmul.4s v7, v2, v4
    WORD $0x6e24dc68          // fmul.4s v8, v3, v4

    WORD $0x6e25d4a9          // faddp.4s v9, v5, v5
    WORD $0x7e30d920          // faddp.2s s0, v9

    WORD $0x6e26d4ca          // faddp.4s v10, v6, v6
    WORD $0x7e30d941          // faddp.2s s1, v10

    WORD $0x6e27d4eb          // faddp.4s v11, v7, v7
    WORD $0x7e30d962          // faddp.2s s2, v11

    WORD $0x6e28d50c          // faddp.4s v12, v8, v8
    WORD $0x7e30d983          // faddp.2s s3, v12

    FMOVS F0, ret_0+80(FP)
    FMOVS F1, ret_1+84(FP)
    FMOVS F2, ret_2+88(FP)
    FMOVS F3, ret_3+92(FP)
    RET
