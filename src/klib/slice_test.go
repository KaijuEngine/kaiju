/******************************************************************************/
/* slice_test.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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
