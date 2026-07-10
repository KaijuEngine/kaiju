/******************************************************************************/
/* virtual_layout_linear_test.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"testing"

	"kaijuengine.com/matrix"
)

// These golden tests pin the default LinearVertical strategy to the exact
// geometry the pre-seam VirtualList produced: cells at x=0 with y/height from the
// height model, content height == model total, and a visible set identical to the
// already-tested virtualWindow. This is the backward-compatibility gate that
// guarantees TextArea (the only engine consumer) is unaffected.

func collectVisible(l VirtualLayout, scroll, vp matrix.Vec2, overscan int) []int {
	var got []int
	l.VisibleAt(scroll, vp, overscan, func(i int) { got = append(got, i) })
	return got
}

func windowSlice(m virtualHeightModel, n, overscan int, vpH, scrollY matrix.Float) []int {
	first, last := virtualWindow(m, n, overscan, vpH, scrollY)
	var out []int
	for i := first; i <= last; i++ {
		out = append(out, i)
	}
	return out
}

func assertIntSlice(t *testing.T, got, want []int, name string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: len = %d %v, want %d %v", name, len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %d, want %d (%v vs %v)", name, i, got[i], want[i], got, want)
		}
	}
}

func TestLinearVerticalGoldenFixed(t *testing.T) {
	const h matrix.Float = 20
	const n = 1000

	ref := newFixedHeightModel(h)
	ref.setCount(n)

	l := newLinearVertical(virtualListDefaultEstimate)
	l.setFixed(h)
	l.SetCount(n)

	if l.Axis() != VirtualAxisVertical {
		t.Fatalf("axis = %d, want vertical", l.Axis())
	}
	assertEqualI(t, l.Count(), n, "count")

	vp := matrix.Vec2{300, 200}
	cs := l.ContentSize(vp)
	assertEqualF(t, cs.X(), 300, "contentW")
	assertEqualF(t, cs.Y(), ref.total(), "contentH")

	for _, i := range []int{0, 1, 7, 499, 999} {
		r := l.RectOf(i, vp)
		assertEqualF(t, r.X, 0, "rectX")
		assertEqualF(t, r.Y, ref.offsetOf(i), "rectY")
		assertEqualF(t, r.W, 300, "rectW")
		assertEqualF(t, r.H, ref.heightOf(i), "rectH")
	}

	// VisibleAt must equal virtualWindow for a spread of scroll/overscan.
	cases := []struct {
		scrollY  matrix.Float
		overscan int
	}{
		{0, 0}, {0, 2}, {1000, 2}, {ref.total() - 200, 2}, {123.5, 4},
	}
	for _, c := range cases {
		got := collectVisible(l, matrix.Vec2{0, c.scrollY}, vp, c.overscan)
		want := windowSlice(ref, n, c.overscan, vp.Y(), c.scrollY)
		assertIntSlice(t, got, want, "visible(fixed)")
	}
}

func TestLinearVerticalGoldenVariable(t *testing.T) {
	const n = 500
	fn := func(i int) matrix.Float { return matrix.Float(10 + (i%5)*7) }

	ref := newPrefixHeightModel(virtualListDefaultEstimate)
	ref.setCount(n)
	for i := range n {
		ref.setHeight(i, fn(i))
	}

	l := newLinearVertical(virtualListDefaultEstimate)
	l.setVariable(fn)
	l.SetCount(n)
	l.Remeasure()

	vp := matrix.Vec2{260, 180}
	assertEqualF(t, l.ContentSize(vp).Y(), ref.total(), "contentH")

	for _, i := range []int{0, 3, 4, 5, 250, 499} {
		r := l.RectOf(i, vp)
		assertEqualF(t, r.X, 0, "rectX")
		assertEqualF(t, r.Y, ref.offsetOf(i), "rectY")
		assertEqualF(t, r.H, ref.heightOf(i), "rectH")
	}

	for _, scrollY := range []matrix.Float{0, 200, 1500, ref.total() - 180} {
		got := collectVisible(l, matrix.Vec2{0, scrollY}, vp, 2)
		want := windowSlice(ref, n, 2, vp.Y(), scrollY)
		assertIntSlice(t, got, want, "visible(variable)")
	}

	// Invalidate a single row mirrors a model setHeight + reflow.
	l.Invalidate(0) // re-pulls fn(0) == 10, unchanged here
	assertEqualF(t, l.RectOf(1, vp).Y, ref.offsetOf(1), "rectY after invalidate")
}

func TestLinearVerticalMinContentWidth(t *testing.T) {
	l := newLinearVertical(virtualListDefaultEstimate)
	l.setFixed(20)
	l.SetCount(10)

	vp := matrix.Vec2{300, 200}
	assertEqualF(t, l.ContentSize(vp).X(), 300, "contentW default")
	assertEqualF(t, l.RectOf(0, vp).W, 300, "rectW default")

	l.setMinContentW(500)
	assertEqualF(t, l.ContentSize(vp).X(), 500, "contentW widened")
	assertEqualF(t, l.RectOf(0, vp).W, 500, "rectW widened")

	// A wider viewport than minContentW tracks the viewport.
	wide := matrix.Vec2{800, 200}
	assertEqualF(t, l.ContentSize(wide).X(), 800, "contentW tracks wide viewport")
}

func TestLinearVerticalEmpty(t *testing.T) {
	l := newLinearVertical(virtualListDefaultEstimate)
	l.setFixed(20)
	l.SetCount(0)
	vp := matrix.Vec2{300, 200}
	assertEqualF(t, l.ContentSize(vp).Y(), 0, "empty contentH")
	got := collectVisible(l, matrix.Vec2{0, 0}, vp, 2)
	if len(got) != 0 {
		t.Fatalf("empty visible = %v, want none", got)
	}
}
