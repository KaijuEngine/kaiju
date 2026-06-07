/******************************************************************************/
/* virtual_list_test.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "testing"

func TestVirtualWindowFixed(t *testing.T) {
	m := newFixedHeightModel(20)
	m.setCount(100_000)

	// Viewport 200px (10 rows) at the top, overscan 0.
	first, last := virtualWindow(m, m.count(), 0, 200, 0)
	assertEqualI(t, first, 0, "first")
	assertEqualI(t, last, 10, "last") // indexAt(200) == 10

	// Overscan expands both ends and clamps at 0.
	first, last = virtualWindow(m, m.count(), 2, 200, 0)
	assertEqualI(t, first, 0, "first overscan")
	assertEqualI(t, last, 12, "last overscan")

	// Scrolled into the middle: window tracks the offset, count stays bounded.
	first, last = virtualWindow(m, m.count(), 2, 200, 1000) // indexAt(1000)=50
	assertEqualI(t, first, 48, "first mid")
	assertEqualI(t, last, 62, "last mid") // indexAt(1200)=60 +2
	if last-first > 15 {
		t.Fatalf("window size = %d, want <= 15 (must stay viewport-sized)", last-first)
	}

	// Scrolled to the very bottom: last clamps to n-1.
	bottom := m.total() - 200
	first, last = virtualWindow(m, m.count(), 2, 200, bottom)
	assertEqualI(t, last, 99_999, "last bottom")
	if first >= last {
		t.Fatalf("first %d >= last %d at bottom", first, last)
	}
}

func TestVirtualWindowEmpty(t *testing.T) {
	m := newFixedHeightModel(20)
	first, last := virtualWindow(m, 0, 2, 200, 0)
	assertEqualI(t, first, 0, "first")
	assertEqualI(t, last, -1, "last (empty -> last < first)")
}

func TestVirtualWindowShorterThanViewport(t *testing.T) {
	m := newFixedHeightModel(20)
	m.setCount(3) // 60px of content in a 200px viewport
	first, last := virtualWindow(m, m.count(), 2, 200, 0)
	assertEqualI(t, first, 0, "first")
	assertEqualI(t, last, 2, "last (clamps to final row)")
}

func TestVirtualWindowVariableHeights(t *testing.T) {
	m := newPrefixHeightModel(10)
	m.setCount(5)
	m.setHeight(0, 30) // tops: 0,30,40,140,150 ; total 160
	m.setHeight(2, 100)

	// Viewport [35,135) -> rows 1 and 2, overscan 0.
	first, last := virtualWindow(m, m.count(), 0, 100, 35)
	assertEqualI(t, first, 1, "first") // indexAt(35)=1
	assertEqualI(t, last, 2, "last")   // indexAt(135)=2
}
