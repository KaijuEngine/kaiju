/******************************************************************************/
/* virtual_list_height_test.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "testing"

// Shared assertion helpers for the virtual-list / document tests. The engine
// package's tests use the standard library only (no testify).

func assertEqualI(t *testing.T, got, want int, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %d, want %d", name, got, want)
	}
}

func assertEqualF(t *testing.T, got, want float32, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func assertEqualStr(t *testing.T, got, want string, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %q, want %q", name, got, want)
	}
}

func assertEqualPos(t *testing.T, got, want textPos, name string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

func assertTrue(t *testing.T, cond bool, name string) {
	t.Helper()
	if !cond {
		t.Fatalf("%s: expected true", name)
	}
}

func TestFixedHeightModel(t *testing.T) {
	m := newFixedHeightModel(20)
	m.setCount(100)

	assertEqualI(t, m.count(), 100, "count")
	assertEqualF(t, m.total(), 2000, "total")
	assertEqualF(t, m.offsetOf(0), 0, "offsetOf(0)")
	assertEqualF(t, m.offsetOf(10), 200, "offsetOf(10)")
	assertEqualF(t, m.heightOf(0), 20, "heightOf(0)")

	assertEqualI(t, m.indexAt(0), 0, "indexAt(0)")
	assertEqualI(t, m.indexAt(19), 0, "indexAt(19)")
	assertEqualI(t, m.indexAt(20), 1, "indexAt(20)")
	assertEqualI(t, m.indexAt(119), 5, "indexAt(119)")
	// Past the end clamps to last row.
	assertEqualI(t, m.indexAt(10_000), 99, "indexAt(past end)")
	// Negative scroll clamps to first row.
	assertEqualI(t, m.indexAt(-50), 0, "indexAt(negative)")
}

func TestFixedHeightModelEmptyAndZero(t *testing.T) {
	m := newFixedHeightModel(20)
	assertEqualF(t, m.total(), 0, "empty total")
	assertEqualI(t, m.indexAt(100), 0, "empty indexAt")

	z := newFixedHeightModel(0)
	z.setCount(10)
	assertEqualF(t, z.total(), 0, "zero-height total")
	assertEqualI(t, z.indexAt(50), 0, "zero-height indexAt")

	m.setCount(-5)
	assertEqualI(t, m.count(), 0, "negative count clamps to 0")
}

func TestFixedHeightModelRowHeightChange(t *testing.T) {
	m := newFixedHeightModel(10)
	m.setCount(50)
	assertEqualF(t, m.total(), 500, "total before change")
	m.setRowHeight(25)
	assertEqualF(t, m.total(), 1250, "total after change")
	assertEqualF(t, m.offsetOf(10), 250, "offsetOf after change")
}

func TestPrefixHeightModelDefaults(t *testing.T) {
	m := newPrefixHeightModel(15)
	m.setCount(10)

	assertEqualI(t, m.count(), 10, "count")
	// Unmeasured rows use the estimate.
	assertEqualF(t, m.total(), 150, "total")
	assertEqualF(t, m.offsetOf(0), 0, "offsetOf(0)")
	assertEqualF(t, m.offsetOf(3), 45, "offsetOf(3)")
	assertEqualF(t, m.heightOf(0), 15, "heightOf(0)")
}

func TestPrefixHeightModelMeasuredHeights(t *testing.T) {
	m := newPrefixHeightModel(10)
	m.setCount(5)
	assertEqualF(t, m.total(), 50, "initial total")

	// Measure rows out of order; total + offsets reflect the measurements.
	m.setHeight(0, 30)
	m.setHeight(2, 100)
	// rows: [30, 10, 100, 10, 10] => total 160
	assertEqualF(t, m.total(), 160, "total after measure")
	assertEqualF(t, m.offsetOf(0), 0, "offsetOf(0)")
	assertEqualF(t, m.offsetOf(1), 30, "offsetOf(1)")
	assertEqualF(t, m.offsetOf(2), 40, "offsetOf(2)")
	assertEqualF(t, m.offsetOf(3), 140, "offsetOf(3)")
	assertEqualF(t, m.offsetOf(4), 150, "offsetOf(4)")
}

func TestPrefixHeightModelIndexAt(t *testing.T) {
	m := newPrefixHeightModel(10)
	m.setCount(5)
	m.setHeight(0, 30) // rows [30,10,100,10,10] offsets 0,30,40,140,150 total 160
	m.setHeight(2, 100)

	assertEqualI(t, m.indexAt(0), 0, "indexAt(0)")
	assertEqualI(t, m.indexAt(29), 0, "indexAt(29)")
	assertEqualI(t, m.indexAt(30), 1, "indexAt(30)")
	assertEqualI(t, m.indexAt(39), 1, "indexAt(39)")
	assertEqualI(t, m.indexAt(40), 2, "indexAt(40)")
	assertEqualI(t, m.indexAt(139), 2, "indexAt(139)")
	assertEqualI(t, m.indexAt(140), 3, "indexAt(140)")
	assertEqualI(t, m.indexAt(150), 4, "indexAt(150)")
	assertEqualI(t, m.indexAt(159), 4, "indexAt(159)")
	// Past the end clamps to last row.
	assertEqualI(t, m.indexAt(10_000), 4, "indexAt(past end)")
	assertEqualI(t, m.indexAt(-5), 0, "indexAt(negative)")
}

func TestPrefixHeightModelShrinkGrowReset(t *testing.T) {
	m := newPrefixHeightModel(10)
	m.setCount(5)
	m.setHeight(1, 50)
	assertEqualF(t, m.total(), 90, "total after measure")

	// Shrinking keeps remaining measurements.
	m.setCount(2)
	assertEqualI(t, m.count(), 2, "count after shrink")
	assertEqualF(t, m.total(), 60, "total after shrink") // [10,50]

	// Growing fills with estimate.
	m.setCount(4)
	assertEqualF(t, m.total(), 80, "total after grow") // [10,50,10,10]

	// reset drops measurements back to estimate.
	m.reset()
	assertEqualF(t, m.total(), 40, "total after reset")
}

func TestPrefixHeightModelEmpty(t *testing.T) {
	m := newPrefixHeightModel(10)
	assertEqualI(t, m.count(), 0, "count")
	assertEqualF(t, m.total(), 0, "total")
	assertEqualI(t, m.indexAt(100), 0, "indexAt")
	assertEqualF(t, m.offsetOf(0), 0, "offsetOf(0)")
}
