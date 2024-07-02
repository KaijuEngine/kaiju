/******************************************************************************/
/* bitmap_test.go                                                             */
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

package bitmap

import (
	"testing"
)

func TestCheck(t *testing.T) {
	bits := New(8)
	bits.Set(0)
	bits.Set(2)
	bits.Set(4)
	expected := []bool{true, false, true, false, true, false, false, false}
	for i, v := range expected {
		if legacyCheck(bits, i) != v {
			t.Error("failed to correctly check the bits using Go")
		}
		if Check(bits, i) != v {
			t.Error("failed to correctly check the bits using Assembly")
		}
	}
}

func TestCount(t *testing.T) {
	bits := New(16)
	sets := []int{0, 2, 4, 6, 7, 10}
	for i := range sets {
		bits.Set(sets[i])
	}
	if legacyCount(bits) != len(sets) {
		t.Error("failed to correctly count the bits using Go")
	}
	if Count(bits) != len(sets) {
		t.Error("failed to correctly count the bits using Assembly")
	}
}
func BenchmarkCheckGo(b *testing.B) {
	bits := New(8)
	bits.Set(0)
	bits.Set(2)
	bits.Set(4)
	k := 0
	for i := 0; i < b.N; i++ {
		legacyCheck(bits, k)
		k = (k + 1) % 8
	}
}

func BenchmarkCheckAmd64(b *testing.B) {
	bits := New(8)
	bits.Set(0)
	bits.Set(2)
	bits.Set(4)
	k := 0
	for i := 0; i < b.N; i++ {
		Check(bits, k)
		k = (k + 1) % 8
	}
}

func BenchmarkCountGo(b *testing.B) {
	bits := New(16)
	sets := []int{0, 2, 4, 6, 7, 10}
	for i := range sets {
		bits.Set(sets[i])
	}
	for i := 0; i < b.N; i++ {
		legacyCount(bits)
	}
}

func BenchmarkCountAmd64(b *testing.B) {
	bits := New(16)
	sets := []int{0, 2, 4, 6, 7, 10}
	for i := range sets {
		bits.Set(sets[i])
	}
	for i := 0; i < b.N; i++ {
		Count(bits)
	}
}

// Legacy functions for benchmarking
func legacyCheck(b Bitmap, index int) bool {
	return (b[index/bitsInByte] & (0x01 << (index % bitsInByte))) != 0
}

// Count returns the number of bits that are true.
func legacyCount(b Bitmap) int {
	count := 0
	length := len(b) * bitsInByte
	for i := 0; i < length; i++ {
		if Check(b, i) {
			count++
		}
	}
	return count
}
