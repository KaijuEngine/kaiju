/******************************************************************************/
/* slice_test.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"encoding/binary"
	"math"
	"math/rand/v2"
	"slices"
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

func TestShuffleNilRNG(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	Shuffle(slice[:3], nil)
	if len(slice) != 5 {
		t.Fatalf("len(slice) = %d, expected 5", len(slice))
	}
	if !SliceContainsAll(slice, 1, 2, 3, 4, 5) {
		t.Errorf("Shuffle with nil RNG lost elements: %v", slice)
	}
	// A nil slice and a single-element slice must not change or panic.
	Shuffle([]int{}, nil)
	Shuffle([]int{7}, nil)
}

func TestShuffleRandomKeepsElements(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	ShuffleRandom(slice)
	if len(slice) != 5 || !SliceContainsAll(slice, 1, 2, 3, 4, 5) {
		t.Errorf("ShuffleRandom lost elements: %v", slice)
	}
	ShuffleRandom([]int{})
	ShuffleRandom([]int{9})
}

func TestSliceContains(t *testing.T) {
	if !Contains([]string{"a", "b", "c"}, "b") {
		t.Error("expected Contains to find \"b\"")
	}
	if Contains([]string{"a", "b", "c"}, "z") {
		t.Error("did not expect Contains to find \"z\"")
	}
	if Contains([]int{}, 1) {
		t.Error("did not expect Contains to find an item in an empty slice")
	}
}

func TestAppendUnique(t *testing.T) {
	slice := AppendUnique([]int{1, 2}, 2, 3, 3, 4, 1)
	compare := []int{1, 2, 3, 4}
	if len(slice) != len(compare) {
		t.Fatalf("len(slice) = %d, expected %d", len(slice), len(compare))
	}
	for i := range compare {
		if slice[i] != compare[i] {
			t.Errorf("slice[%d] = %d, expected %d", i, slice[i], compare[i])
		}
	}
	// Appending no values returns the original slice unchanged.
	orig := []int{7}
	got := AppendUnique(orig)
	if len(got) != 1 || got[0] != 7 {
		t.Errorf("AppendUnique with no values changed the slice: %v", got)
	}
}

func TestByteSliceToFloat32Slice(t *testing.T) {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:], math.Float32bits(1.5))
	binary.LittleEndian.PutUint32(data[4:], math.Float32bits(-2.25))
	binary.LittleEndian.PutUint32(data[8:], math.Float32bits(0.0))
	binary.LittleEndian.PutUint32(data[12:], math.Float32bits(3.5))
	floats := ByteSliceToFloat32Slice(data)
	compare := []float32{1.5, -2.25, 0.0, 3.5}
	if len(floats) != len(compare) {
		t.Fatalf("len(floats) = %d, expected %d", len(floats), len(compare))
	}
	for i := range compare {
		if floats[i] != compare[i] {
			t.Errorf("floats[%d] = %v, expected %v", i, floats[i], compare[i])
		}
	}
	// Length and capacity must both be trimmed to an exact float32 count.
	if cap(floats) != len(floats) {
		t.Errorf("cap(floats) = %d, expected %d", cap(floats), len(floats))
	}
	if got := ByteSliceToFloat32Slice(nil); len(got) != 0 {
		t.Errorf("ByteSliceToFloat32Slice(nil) len = %d, expected 0", len(got))
	}
}

func TestByteSliceToUInt16Slice(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint16(data[0:], 10)
	binary.LittleEndian.PutUint16(data[2:], 20)
	binary.LittleEndian.PutUint16(data[4:], 30)
	binary.LittleEndian.PutUint16(data[6:], 65535)
	uints := ByteSliceToUInt16Slice(data)
	compare := []uint16{10, 20, 30, 65535}
	if len(uints) != len(compare) {
		t.Fatalf("len(uints) = %d, expected %d", len(uints), len(compare))
	}
	for i := range compare {
		if uints[i] != compare[i] {
			t.Errorf("uints[%d] = %d, expected %d", i, uints[i], compare[i])
		}
	}
	if cap(uints) != len(uints) {
		t.Errorf("cap(uints) = %d, expected %d", cap(uints), len(uints))
	}
	if got := ByteSliceToUInt16Slice(nil); len(got) != 0 {
		t.Errorf("ByteSliceToUInt16Slice(nil) len = %d, expected 0", len(got))
	}
}

func TestRemoveNils(t *testing.T) {
	a, b, c := 1, 2, 3
	slice := []*int{&a, nil, &b, nil, &c}
	result := RemoveNils(slice)
	if len(result) != 3 {
		t.Fatalf("len(result) = %d, expected 3", len(result))
	}
	if result[0] != &a || result[1] != &b || result[2] != &c {
		t.Errorf("RemoveNils returned unexpected pointers: %v", result)
	}
	if len(RemoveNils([]*int{nil, nil})) != 0 {
		t.Error("expected RemoveNils to return an empty slice for all nils")
	}
	if len(RemoveNils([]*int{})) != 0 {
		t.Error("expected RemoveNils to return an empty slice for an empty input")
	}
}

func TestSliceMove(t *testing.T) {
	cases := []struct {
		name string
		from int
		to   int
		want []int
	}{
		{"same index no-op", 2, 2, []int{0, 1, 2, 3, 4}},
		{"adjacent forward", 1, 2, []int{0, 2, 1, 3, 4}},
		{"adjacent backward", 2, 1, []int{0, 2, 1, 3, 4}},
		{"forward over gap", 1, 3, []int{0, 2, 3, 1, 4}},
		{"backward over gap", 3, 1, []int{0, 3, 1, 2, 4}},
		{"from first to last", 0, 4, []int{1, 2, 3, 4, 0}},
		{"from last to first", 4, 0, []int{4, 0, 1, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := []int{0, 1, 2, 3, 4}
			SliceMove(s, tc.from, tc.to)
			if !slices.Equal(s, tc.want) {
				t.Errorf("SliceMove(0,1,2,3,4, %d, %d) = %v, expected %v", tc.from, tc.to, s, tc.want)
			}
		})
	}
}

func TestSliceSetCap(t *testing.T) {
	s := make([]int, 3, 3)
	grown := SliceSetCap(s, 10)
	if len(grown) != 3 {
		t.Errorf("len = %d, expected 3", len(grown))
	}
	if cap(grown) < 10 {
		t.Errorf("cap = %d, expected at least 10", cap(grown))
	}
	if !slices.Equal(grown[:3], []int{0, 0, 0}) {
		t.Errorf("SliceSetCap changed existing elements: %v", grown[:3])
	}
	same := SliceSetCap(s, 2)
	if same != nil && (len(same) != 3 || cap(same) != 3) {
		t.Errorf("SliceSetCap with amount below len changed the slice: %v", same)
	}
}

func TestSliceSetLen(t *testing.T) {
	// Growing beyond capacity triggers SliceSetCap internally.
	grown := SliceSetLen([]int{1, 2}, 6)
	if len(grown) != 6 {
		t.Fatalf("len = %d, expected 6", len(grown))
	}
	if !slices.Equal(grown[:2], []int{1, 2}) {
		t.Errorf("SliceSetLen lost original elements: %v", grown[:2])
	}
	for i := 2; i < 6; i++ {
		if grown[i] != 0 {
			t.Errorf("grown[%d] = %d, expected 0", i, grown[i])
		}
	}
	// Growing within the existing capacity.
	short := make([]int, 2, 8)
	short[0], short[1] = 5, 6
	grownInPlace := SliceSetLen(short, 4)
	if len(grownInPlace) != 4 {
		t.Fatalf("len = %d, expected 4", len(grownInPlace))
	}
	if !slices.Equal(grownInPlace[:2], []int{5, 6}) {
		t.Errorf("SliceSetLen lost elements: %v", grownInPlace[:2])
	}
	// Shrinking re-slices without touching content.
	shrunk := SliceSetLen(grown, 2)
	if len(shrunk) != 2 {
		t.Fatalf("len = %d, expected 2", len(shrunk))
	}
	if !slices.Equal(shrunk, []int{1, 2}) {
		t.Errorf("SliceSetLen shrink changed elements: %v", shrunk)
	}
	// Truncating to zero length.
	zero := SliceSetLen(grown, 0)
	if len(zero) != 0 {
		t.Errorf("len = %d, expected 0", len(zero))
	}
}

func TestSlicesAreTheSame(t *testing.T) {
	if !SlicesAreTheSame([]int{1, 2, 3}, []int{1, 2, 3}) {
		t.Error("expected equal slices to be the same")
	}
	if SlicesAreTheSame([]int{1, 2, 3}, []int{1, 2, 4}) {
		t.Error("did not expect differing slices to be the same")
	}
	if SlicesAreTheSame([]int{1, 2}, []int{1, 2, 3}) {
		t.Error("did not expect slices of different lengths to be the same")
	}
	if !SlicesAreTheSame([]string{}, []string{}) {
		t.Error("expected two empty slices to be the same")
	}
}

func TestSlicesRemoveElement(t *testing.T) {
	got := SlicesRemoveElement([]int{1, 3, 2, 3, 3, 4}, 3)
	if !slices.Equal(got, []int{1, 2, 4}) {
		t.Errorf("SlicesRemoveElement = %v, expected [1 2 4]", got)
	}
	// Removing an element that is not present leaves the slice untouched.
	unchanged := SlicesRemoveElement([]int{1, 2}, 9)
	if !slices.Equal(unchanged, []int{1, 2}) {
		t.Errorf("SlicesRemoveElement = %v, expected [1 2]", unchanged)
	}
	empty := SlicesRemoveElement([]int{}, 1)
	if len(empty) != 0 {
		t.Errorf("SlicesRemoveElement on empty = %v, expected []", empty)
	}
}

func TestStringsContainsCaseInsensitive(t *testing.T) {
	if !StringsContainsCaseInsensitive([]string{"Hello", "World"}, "hello") {
		t.Error("expected case-insensitive match for \"hello\"")
	}
	if !StringsContainsCaseInsensitive([]string{"Hello", "World"}, "WORLD") {
		t.Error("expected case-insensitive match for \"WORLD\"")
	}
	if StringsContainsCaseInsensitive([]string{"Hello", "World"}, "nope") {
		t.Error("did not expect a match for \"nope\"")
	}
	if StringsContainsCaseInsensitive([]string{}, "x") {
		t.Error("did not expect a match in an empty slice")
	}
}

func TestWipeSlice(t *testing.T) {
	s := []int{1, 2, 3}
	wiped := WipeSlice(s)
	if len(wiped) != 0 {
		t.Errorf("len = %d, expected 0", len(wiped))
	}
	if cap(wiped) != 3 {
		t.Errorf("cap = %d, expected 3", cap(wiped))
	}
	// WipeSlice must also have cleared the underlying backing array.
	for i := range s {
		if s[i] != 0 {
			t.Errorf("backing array element %d = %d, expected 0", i, s[i])
		}
	}
	// An already-empty slice comes back unchanged.
	got := WipeSlice([]int{})
	if got == nil || len(got) != 0 {
		t.Errorf("WipeSlice(empty) = %v, expected empty non-nil slice", got)
	}
}

func TestRemakeSlice(t *testing.T) {
	s := []int{1, 2, 3}
	remade := RemakeSlice(s)
	if len(remade) != 0 {
		t.Errorf("len = %d, expected 0", len(remade))
	}
	if cap(remade) != 3 {
		t.Errorf("cap = %d, expected 3", cap(remade))
	}
	empty := RemakeSlice([]int{})
	if empty == nil || len(empty) != 0 {
		t.Errorf("RemakeSlice(empty) = %v, expected empty non-nil slice", empty)
	}
}

func TestExtractFromSlice(t *testing.T) {
	s := []int{10, 20, 30}
	res := ExtractFromSlice(s, func(idx int) string {
		return string(rune('a'+idx))
	})
	if !slices.Equal(res, []string{"a", "b", "c"}) {
		t.Errorf("ExtractFromSlice = %v, expected [a b c]", res)
	}
	// The transform can map values from the slice itself, and empty input
	// yields an empty non-nil result.
	doubled := ExtractFromSlice(s, func(idx int) int { return s[idx] * 2 })
	if !slices.Equal(doubled, []int{20, 40, 60}) {
		t.Errorf("ExtractFromSlice doubled = %v, expected [20 40 60]", doubled)
	}
	empty := ExtractFromSlice([]int{}, func(idx int) int { return idx })
	if empty == nil || len(empty) != 0 {
		t.Errorf("ExtractFromSlice(empty) = %v, expected empty non-nil slice", empty)
	}
}

// SliceContainsAll is a small test helper that reports whether haystack
// contains every item exactly once.
func SliceContainsAll[T comparable](haystack []T, items ...T) bool {
	for _, item := range items {
		if !Contains(haystack, item) {
			return false
		}
	}
	return len(haystack) == len(items)
}
