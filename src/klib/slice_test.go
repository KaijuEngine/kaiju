/******************************************************************************/
/* slice_test.go                                                              */
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

package klib

import (
	"math/rand/v2"
	"testing"
)

func TestRemoveUnordered(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	slice = RemoveUnordered(slice, 2)
	compare := []int{1, 2, 5, 4}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

func TestShuffle(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	Shuffle(slice, rand.New(rand.NewPCG(0, 0)))
	compare := []int{1, 3, 4, 5, 2}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

func TestShuffleFront(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	Shuffle(slice[:3], rand.New(rand.NewPCG(0, 0)))
	compare := []int{3, 2, 1, 4, 5}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

func TestShuffleBack(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	Shuffle(slice[2:], rand.New(rand.NewPCG(0, 0)))
	compare := []int{1, 2, 5, 4, 3}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

func TestShuffleEmpty(t *testing.T) {
	slice := make([]int, 0)
	Shuffle(slice, rand.New(rand.NewPCG(0, 0)))
	compare := make([]int, 0)
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

func TestShuffleOne(t *testing.T) {
	slice := []int{1}
	Shuffle(slice, rand.New(rand.NewPCG(0, 0)))
	compare := []int{1}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

func TestShuffleMiddle(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	Shuffle(slice[1:4], rand.New(rand.NewPCG(0, 0)))
	compare := []int{1, 4, 3, 2, 5}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}
