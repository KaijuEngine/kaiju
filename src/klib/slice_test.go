package klib

import (
	"math/rand"
	"testing"
)

func TestRemoveOrdered(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	slice = RemoveOrdered(slice, 2)
	compare := []int{1, 2, 4, 5}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}

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
	Shuffle(slice, rand.New(rand.NewSource(0)))
	compare := []int{4, 1, 2, 3, 5}
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
	Shuffle(slice[:3], rand.New(rand.NewSource(0)))
	compare := []int{2, 3, 1, 4, 5}
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
	Shuffle(slice[2:], rand.New(rand.NewSource(0)))
	compare := []int{1, 2, 4, 5, 3}
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
	Shuffle(slice, rand.New(rand.NewSource(0)))
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
	Shuffle(slice, rand.New(rand.NewSource(0)))
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
	Shuffle(slice[1:4], rand.New(rand.NewSource(0)))
	compare := []int{1, 3, 4, 2, 5}
	if len(slice) != len(compare) {
		t.Errorf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := 0; i < len(slice); i++ {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
}
