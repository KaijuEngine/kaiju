/******************************************************************************/
/* bitmap.amd64.s                                                             */
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

// func Check(b Bitmap, index int) bool
TEXT   ·Check(SB),NOSPLIT,$0-32
	MOVQ b+0(FP), DX          // Head address to slice data 
	MOVBLZX index+24(FP), CX  // index
	SHRL $3, CX               // Divide by 8 (8-bits in byte)
	ADDQ CX, DX               // Offset into bit map
	MOVB (DX), AX             // Read the specified byte
	MOVL index+24(FP), CX     // index
	ANDL $7, CX
	BTW CX, AX
	//SETCS AL                // Typically in Go, a boolean is returned in
	//MOVB AL, index+32(FP)   // AX but that doesn't seem to be the case
	SETCS index+32(FP)        // for embedded assembly in Go
	RET

// func CountASM(b Bitmap) int
TEXT   ·CountASM(SB),NOSPLIT,$0-28
	MOVQ b+0(FP), DX       // Head address to slice data 
	MOVWLZX b+8(FP), CX    // Byte length
	XORQ R8, R8            // 0
	XORQ AX, AX            // 0
	MOVL AX, index+24(FP)  // Return int
count:
	MOVB (DX), R8          // Read the next byte of the slice
	INCQ DX                // Move the pointer to the next byte address
	POPCNTW R8, AX         // Population count of 1s in word
	ADDL AX, index+24(FP)  // Add the calculated pop count to return
	SUBL $1, CX            // Decrement our counter
	JNE count              // If our counter is not 0, continue loop
exit:
	RET

DATA poptab<>+0x00(SB)/4, $0x02010100
DATA poptab<>+0x04(SB)/4, $0x03020201
DATA poptab<>+0x08(SB)/4, $0x03020201
DATA poptab<>+0x0C(SB)/4, $0x04030302
DATA poptab<>+0x10(SB)/4, $0x03020201
DATA poptab<>+0x14(SB)/4, $0x04030302
DATA poptab<>+0x18(SB)/4, $0x04030302
DATA poptab<>+0x1C(SB)/4, $0x05040403
DATA poptab<>+0x20(SB)/4, $0x03020201
DATA poptab<>+0x24(SB)/4, $0x04030302
DATA poptab<>+0x28(SB)/4, $0x04030302
DATA poptab<>+0x2C(SB)/4, $0x05040403
DATA poptab<>+0x30(SB)/4, $0x04030302
DATA poptab<>+0x34(SB)/4, $0x05040403
DATA poptab<>+0x38(SB)/4, $0x05040403
DATA poptab<>+0x3C(SB)/4, $0x06050504
DATA poptab<>+0x40(SB)/4, $0x03020201
DATA poptab<>+0x44(SB)/4, $0x04030302
DATA poptab<>+0x48(SB)/4, $0x04030302
DATA poptab<>+0x4C(SB)/4, $0x05040403
DATA poptab<>+0x50(SB)/4, $0x04030302
DATA poptab<>+0x54(SB)/4, $0x05040403
DATA poptab<>+0x58(SB)/4, $0x05040403
DATA poptab<>+0x5C(SB)/4, $0x06050504
DATA poptab<>+0x60(SB)/4, $0x04030302
DATA poptab<>+0x64(SB)/4, $0x05040403
DATA poptab<>+0x68(SB)/4, $0x05040403
DATA poptab<>+0x6C(SB)/4, $0x06050504
DATA poptab<>+0x70(SB)/4, $0x05040403
DATA poptab<>+0x74(SB)/4, $0x06050504
DATA poptab<>+0x78(SB)/4, $0x06050504
DATA poptab<>+0x7C(SB)/4, $0x07060605
DATA poptab<>+0x80(SB)/4, $0x03020201
DATA poptab<>+0x84(SB)/4, $0x04030302
DATA poptab<>+0x88(SB)/4, $0x04030302
DATA poptab<>+0x8C(SB)/4, $0x05040403
DATA poptab<>+0x90(SB)/4, $0x04030302
DATA poptab<>+0x94(SB)/4, $0x05040403
DATA poptab<>+0x98(SB)/4, $0x05040403
DATA poptab<>+0x9C(SB)/4, $0x06050504
DATA poptab<>+0xA0(SB)/4, $0x04030302
DATA poptab<>+0xA4(SB)/4, $0x05040403
DATA poptab<>+0xA8(SB)/4, $0x05040403
DATA poptab<>+0xAC(SB)/4, $0x06050504
DATA poptab<>+0xB0(SB)/4, $0x05040403
DATA poptab<>+0xB4(SB)/4, $0x06050504
DATA poptab<>+0xB8(SB)/4, $0x06050504
DATA poptab<>+0xBC(SB)/4, $0x07060605
DATA poptab<>+0xC0(SB)/4, $0x04030302
DATA poptab<>+0xC4(SB)/4, $0x05040403
DATA poptab<>+0xC8(SB)/4, $0x05040403
DATA poptab<>+0xCC(SB)/4, $0x06050504
DATA poptab<>+0xD0(SB)/4, $0x05040403
DATA poptab<>+0xD4(SB)/4, $0x06050504
DATA poptab<>+0xD8(SB)/4, $0x06050504
DATA poptab<>+0xDC(SB)/4, $0x07060605
DATA poptab<>+0xE0(SB)/4, $0x05040403
DATA poptab<>+0xE4(SB)/4, $0x06050504
DATA poptab<>+0xE8(SB)/4, $0x06050504
DATA poptab<>+0xEC(SB)/4, $0x07060605
DATA poptab<>+0xF0(SB)/4, $0x06050504
DATA poptab<>+0xF4(SB)/4, $0x07060605
DATA poptab<>+0xF8(SB)/4, $0x07060605
DATA poptab<>+0xFC(SB)/4, $0x08070706
GLOBL poptab<>(SB), RODATA, $256

// func CountASMUsingTable(b Bitmap) int
TEXT   ·CountASMUsingTable(SB),NOSPLIT,$0-28
	MOVQ b+0(FP), DX       // Head address to slice data 
	MOVL b+8(FP), CX       // Byte length
	MOVL $0, index+24(FP)  // Return int
	XORQ R8, R8
	LEAQ poptab<>(SB), R9
	XORQ AX, AX
count:
	MOVB (DX), R8
	INCQ DX
	MOVB (R9)(R8*1), AX
	ADDL AX, index+24(FP)
	SUBL $1, CX
	JNE count
exit:
	RET
