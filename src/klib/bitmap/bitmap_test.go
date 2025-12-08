/******************************************************************************/
/* bitmap_test.go                                                             */
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

package bitmap

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/KaijuEngine/kaiju/klib"
)

func seededRandomTestSet(maxMapLen int, seed1, seed2 uint64) []int {
	rnd := rand.New(rand.NewPCG(seed1, seed2))
	mapLen := rnd.IntN(maxMapLen) + 1
	onBits := rnd.IntN(mapLen) + 1
	choices := make([]int, onBits)
	for i := range onBits {
		choices[i] = i
	}
	klib.Shuffle(choices, rnd)
	return choices[:onBits]
}

func randomTestSet(maxMapLen int) ([]int, uint64, uint64) {
	seed1 := uint64(time.Now().UnixNano())
	seed2 := uint64(float64(time.Now().UnixNano()) * 0.13)
	return seededRandomTestSet(maxMapLen, seed1, seed2), seed1, seed2
}

func TestCheck(t *testing.T) {
	sets, seed1, seed2 := randomTestSet(64)
	count := slices.Max(sets) + 1
	bits := New(count)
	expected := make([]bool, count)
	for i := range sets {
		expected[sets[i]] = true
		bits.Set(sets[i])
	}
	for i, v := range expected {
		if legacyCheck(bits, i) != v {
			t.Fatalf("[Go] Index %d was expected to be %v but was %v for seed %d:%d", i, v, legacyCheck(bits, i), seed1, seed2)
		}
		if Check(bits, i) != v {
			t.Fatalf("[Asm] Index %d was expected to be %v but was %v for seed %d:%d", i, v, Check(bits, i), seed1, seed2)
		}
	}
}

func TestCount(t *testing.T) {
	//seed1 := int64(1720034999808757000)
	//seed2 := int64(1720034181037542000)
	//sets := seededRandomTestSet(64, seed1, seed2)
	sets, seed1, seed2 := randomTestSet(64)
	bits := New(slices.Max(sets) + 1)
	for i := range sets {
		bits.Set(sets[i])
	}
	if Count(bits) != len(sets) {
		t.Fatalf("[Go] Count was expected to be %d but was %d for seed %d:%d", len(sets), Count(bits), seed1, seed2)
	}
	if legacyCount(bits) != len(sets) {
		t.Fatalf("[Go] Count was expected to be %d but was %d for seed %d:%d", len(sets), legacyCount(bits), seed1, seed2)
	}
	if CountASM(bits) != len(sets) {
		t.Fatalf("[Asm] Count was expected to be %d but was %d for seed %d:%d", len(sets), CountASM(bits), seed1, seed2)
	}
}
func BenchmarkCheckGo(b *testing.B) {
	const last = 22
	sets := []int{0, 2, 4, 6, 7, 10, 13, 15, 17, 18, 19, 20, last}
	bits := New(last + 1)
	expected := make([]bool, last+1)
	for i := range sets {
		expected[sets[i]] = true
		bits.Set(sets[i])
	}
	k := 0
	for i := 0; i < b.N; i++ {
		legacyCheck(bits, k)
		k = (k + 1) % 8
	}
}

func BenchmarkCheckAmd64(b *testing.B) {
	const last = 22
	sets := []int{0, 2, 4, 6, 7, 10, 13, 15, 17, 18, 19, 20, last}
	bits := New(last + 1)
	expected := make([]bool, last+1)
	for i := range sets {
		expected[sets[i]] = true
		bits.Set(sets[i])
	}
	k := 0
	for i := 0; i < b.N; i++ {
		Check(bits, k)
		k = (k + 1) % 8
	}
}

func BenchmarkCountGo(b *testing.B) {
	sets := seededRandomTestSet(64, 99, 93)
	bits := New(slices.Max(sets) + 1)
	for i := range sets {
		bits.Set(sets[i])
	}
	for i := 0; i < b.N; i++ {
		legacyCount(bits)
	}
}

func BenchmarkCountAmd64(b *testing.B) {
	sets := seededRandomTestSet(64, 99, 93)
	bits := New(slices.Max(sets) + 1)
	for i := range sets {
		bits.Set(sets[i])
	}
	for i := 0; i < b.N; i++ {
		CountASM(bits)
	}
}

func BenchmarkFastCountGo(b *testing.B) {
	sets := seededRandomTestSet(64, 99, 93)
	bits := New(slices.Max(sets) + 1)
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
		if legacyCheck(b, i) {
			count++
		}
	}
	return count
}
