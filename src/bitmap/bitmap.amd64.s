/******************************************************************************/
/* bitmap.amd64.s                                                             */
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

// func Check(b Bitmap, index int) bool
TEXT   ·Check(SB),NOSPLIT,$0-32
	MOVQ b+0(FP), DX        // Head address to slice data 
	MOVL index+24(FP), CX   // index
	SHRL $3, CX             // Divide by 8 (8-bits in byte)
	ADDQ CX, DX             // Offset into bit map
	MOVB (DX), AX           // Read the specified byte
	MOVL index+24(FP), CX   // index
	BTW CX, AX
	//SETCS AL              // Typically in Go, a boolean is returned in
	//MOVB AL, index+32(FP) // AX but that doesn't seem to be the case
	SETCS index+32(FP)      // for embedded assembly in Go
	RET

// func Count(b Bitmap) int
TEXT   ·Count(SB),NOSPLIT,$0-28
	MOVQ b+0(FP), DX        // Head address to slice data 
	MOVW b+8(FP), CX		// Byte length
	XORW R8, R8
	MOVL $0, index+24(FP)
count:
	MOVB (DX), R8
	INCQ DX
	POPCNTW R8, R9
	ADDW R9, index+24(FP)
	SUBW $1, CX
	JNE count
	RET
