/******************************************************************************/
/* label_color_range_test.go                                                  */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ui

import (
	"testing"

	"kaijuengine.com/matrix"
)

// TestFindColorRangeAppendsAndReturnsStablePointer pins the storage invariant
// that the old findColorRange implementation violated: each call must append a
// new entry to ld.colorRanges and return a pointer that addresses the
// slice element, not a stack-local copy that would be GC-collected after the
// frame. We exercise appendColorRange directly because findColorRange is a
// method on *Label and requires a host/font cache to fully construct; the
// helper is the exact production code path findColorRange now delegates to.
//
// The test pre-allocates capacity 2 so the second append cannot trigger a
// realloc that would move the backing array — Go's slice contract does not
// guarantee element-pointer stability across grow, so we assert the in-array
// invariant ("pointer addresses a slice element") under conditions where the
// array is fixed.
func TestFindColorRangeAppendsAndReturnsStablePointer(t *testing.T) {
	ld := &labelData{
		colorRanges: make([]colorRange, 0, 2),
		fgColor:     matrix.ColorWhite(),
		bgColor:     matrix.ColorBlack(),
	}

	first := appendColorRange(ld, 0, 5)
	second := appendColorRange(ld, 5, 10)

	if got, want := len(ld.colorRanges), 2; got != want {
		t.Fatalf("len(ld.colorRanges) = %d, want %d", got, want)
	}
	if first != &ld.colorRanges[0] {
		t.Fatalf("first returned pointer %p, want &ld.colorRanges[0] %p", first, &ld.colorRanges[0])
	}
	if second != &ld.colorRanges[1] {
		t.Fatalf("second returned pointer %p, want &ld.colorRanges[1] %p", second, &ld.colorRanges[1])
	}
	// Mutating through the returned pointer must be visible in the slice —
	// proves the pointer is not aliasing a stack-local copy.
	first.hue = matrix.ColorRed()
	if !ld.colorRanges[0].hue.Equals(matrix.ColorRed()) {
		t.Fatalf("mutation through returned pointer not visible in slice: got %+v, want red", ld.colorRanges[0].hue)
	}
	// Bounds must round-trip exactly as supplied.
	if ld.colorRanges[0].start != 0 || ld.colorRanges[0].end != 5 {
		t.Fatalf("first range bounds = (%d,%d), want (0,5)", ld.colorRanges[0].start, ld.colorRanges[0].end)
	}
	if ld.colorRanges[1].start != 5 || ld.colorRanges[1].end != 10 {
		t.Fatalf("second range bounds = (%d,%d), want (5,10)", ld.colorRanges[1].start, ld.colorRanges[1].end)
	}
	// Seeded hues should reflect ld.fgColor / ld.bgColor at append time.
	if !ld.colorRanges[1].hue.Equals(matrix.ColorWhite()) {
		t.Fatalf("second range hue = %+v, want white", ld.colorRanges[1].hue)
	}
	if !ld.colorRanges[1].bgHue.Equals(matrix.ColorBlack()) {
		t.Fatalf("second range bgHue = %+v, want black", ld.colorRanges[1].bgHue)
	}
}

// TestAppendColorRangeReturnsLastSliceElement covers the bare invariant the
// old `return &newRange` violated: regardless of any slice-grow that may
// happen during append, the returned pointer must equal the address of the
// final element of ld.colorRanges *as observed immediately after the call*.
// This is the production guarantee findColorRange callers (ColorRange,
// BoldRange) depend on when they mutate the returned hue/bgHue fields and
// expect the change to persist in the stored range.
func TestAppendColorRangeReturnsLastSliceElement(t *testing.T) {
	ld := &labelData{
		colorRanges: make([]colorRange, 0),
		fgColor:     matrix.ColorWhite(),
		bgColor:     matrix.ColorBlack(),
	}
	for i := 0; i < 5; i++ {
		got := appendColorRange(ld, i, i+1)
		want := &ld.colorRanges[len(ld.colorRanges)-1]
		if got != want {
			t.Fatalf("iteration %d: returned pointer %p, want last-element pointer %p (slice len=%d, cap=%d)",
				i, got, want, len(ld.colorRanges), cap(ld.colorRanges))
		}
		// Mutation through the returned pointer must be visible — the
		// stack-local-pointer defect would lose this write.
		got.hue = matrix.ColorRed()
		if !ld.colorRanges[len(ld.colorRanges)-1].hue.Equals(matrix.ColorRed()) {
			t.Fatalf("iteration %d: mutation through returned pointer not visible in slice", i)
		}
	}
	if got, want := len(ld.colorRanges), 5; got != want {
		t.Fatalf("len(ld.colorRanges) = %d, want %d", got, want)
	}
}

// TestClearColorRangesEmptiesSlice pins the truncation semantics used by
// ClearColorRanges: len drops to zero but the backing array is preserved so
// repeated wipe/repopulate cycles (e.g. per-keystroke syntax rehighlight in
// M4) stay allocation-free. We perform the truncation directly because
// constructing a *Label requires a host wiring that is out of scope for an
// in-package unit test; the production ClearColorRanges performs the exact
// same `ld.colorRanges = ld.colorRanges[:0]` plus an updateColors call.
func TestClearColorRangesEmptiesSlice(t *testing.T) {
	ld := &labelData{
		colorRanges: make([]colorRange, 0, 8),
		fgColor:     matrix.ColorWhite(),
		bgColor:     matrix.ColorBlack(),
	}

	appendColorRange(ld, 0, 5)
	appendColorRange(ld, 5, 10)
	if got, want := len(ld.colorRanges), 2; got != want {
		t.Fatalf("pre-truncate len = %d, want %d", got, want)
	}
	priorCap := cap(ld.colorRanges)

	// Mirror the production ClearColorRanges truncation.
	ld.colorRanges = ld.colorRanges[:0]

	if got := len(ld.colorRanges); got != 0 {
		t.Fatalf("post-truncate len = %d, want 0", got)
	}
	if got := cap(ld.colorRanges); got < 2 {
		t.Fatalf("post-truncate cap = %d, want >= 2 (backing array must be retained)", got)
	}
	if got := cap(ld.colorRanges); got != priorCap {
		t.Fatalf("post-truncate cap = %d, want %d (backing array must be byte-identical)", got, priorCap)
	}

	// Re-appending after truncate must reuse the array — append must not
	// allocate while we are still under capacity.
	reused := appendColorRange(ld, 0, 3)
	if reused != &ld.colorRanges[0] {
		t.Fatalf("post-truncate append returned %p, want &ld.colorRanges[0] %p", reused, &ld.colorRanges[0])
	}
	if got := cap(ld.colorRanges); got != priorCap {
		t.Fatalf("append after truncate grew cap from %d to %d (lost allocation-free reuse)", priorCap, got)
	}
}
