/******************************************************************************/
/* virtual_list_height.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"sort"

	"kaijuengine.com/matrix"
)

// virtualHeightModel maps row indices to vertical offsets and back for a
// VirtualList. It is the piece that lets the list render only the visible
// window: given a scroll offset it answers which row sits at that y, and given
// a row it answers where that row's top is.
//
// Two implementations exist. fixedHeightModel is used when every row is the
// same height (the no-wrap code-editor path) and is pure O(1) arithmetic, which
// is what keeps 100k lines cheap. prefixHeightModel is used when rows have
// individual heights (the wrapped / chat path) and keeps cumulative prefix sums
// so offsetOf stays O(1) and indexAt is a binary search.
type virtualHeightModel interface {
	// setCount sets the number of rows; new rows take a default height.
	setCount(n int)
	count() int
	// total is the full content height (sum of all row heights).
	total() matrix.Float
	// offsetOf is the y of the top edge of row index.
	offsetOf(index int) matrix.Float
	// heightOf is the height of row index.
	heightOf(index int) matrix.Float
	// indexAt is the row whose vertical span contains y (clamped to range).
	indexAt(y matrix.Float) int
	// setHeight records a measured height for a single row. The fixed model
	// ignores it (all rows share one height).
	setHeight(index int, h matrix.Float)
	// reset drops any per-row measurements back to defaults.
	reset()
}

// fixedHeightModel gives every row the same height. All queries are O(1).
type fixedHeightModel struct {
	n         int
	rowHeight matrix.Float
}

func newFixedHeightModel(rowHeight matrix.Float) *fixedHeightModel {
	return &fixedHeightModel{rowHeight: rowHeight}
}

func (m *fixedHeightModel) setRowHeight(h matrix.Float) { m.rowHeight = h }

func (m *fixedHeightModel) setCount(n int) {
	if n < 0 {
		n = 0
	}
	m.n = n
}

func (m *fixedHeightModel) count() int { return m.n }

func (m *fixedHeightModel) total() matrix.Float { return matrix.Float(m.n) * m.rowHeight }

func (m *fixedHeightModel) offsetOf(index int) matrix.Float {
	index = min(max(index, 0), m.n)
	return matrix.Float(index) * m.rowHeight
}

func (m *fixedHeightModel) heightOf(index int) matrix.Float { return m.rowHeight }

func (m *fixedHeightModel) indexAt(y matrix.Float) int {
	if m.rowHeight <= 0 || m.n == 0 {
		return 0
	}
	if y < 0 {
		return 0
	}
	idx := int(y / m.rowHeight)
	return min(max(idx, 0), m.n-1)
}

func (m *fixedHeightModel) setHeight(index int, h matrix.Float) {}

func (m *fixedHeightModel) reset() {}

// prefixHeightModel gives each row its own height. Rows that have not been
// measured yet use estimate, so total height and scrolling work before every
// row has been realized; measured heights are filled in lazily via setHeight as
// rows scroll into view. Cumulative prefix sums (cum[i] = sum of heights[0..i-1])
// are rebuilt lazily from a dirty watermark so a single setHeight does not force
// a full O(n) recompute on every call.
type prefixHeightModel struct {
	estimate  matrix.Float
	heights   []matrix.Float
	cum       []matrix.Float // len == len(heights)+1; cum[i] = sum heights[0..i-1]
	dirtyFrom int            // first index whose cum entry is stale; len(cum) when clean
}

func newPrefixHeightModel(estimate matrix.Float) *prefixHeightModel {
	if estimate <= 0 {
		estimate = 1
	}
	return &prefixHeightModel{estimate: estimate, cum: []matrix.Float{0}, dirtyFrom: 1}
}

func (m *prefixHeightModel) setEstimate(h matrix.Float) {
	if h <= 0 {
		h = 1
	}
	m.estimate = h
}

func (m *prefixHeightModel) setCount(n int) {
	if n < 0 {
		n = 0
	}
	if n == len(m.heights) {
		return
	}
	if n < len(m.heights) {
		m.heights = m.heights[:n]
	} else {
		for len(m.heights) < n {
			m.heights = append(m.heights, m.estimate)
		}
	}
	m.cum = make([]matrix.Float, n+1)
	m.dirtyFrom = 0
}

func (m *prefixHeightModel) count() int { return len(m.heights) }

// ensureCum rebuilds the cumulative sums from dirtyFrom to the end.
func (m *prefixHeightModel) ensureCum() {
	if m.dirtyFrom > len(m.heights) {
		return
	}
	if len(m.cum) != len(m.heights)+1 {
		m.cum = make([]matrix.Float, len(m.heights)+1)
		m.dirtyFrom = 0
	}
	start := max(m.dirtyFrom, 0)
	for i := start; i < len(m.heights); i++ {
		m.cum[i+1] = m.cum[i] + m.heights[i]
	}
	m.dirtyFrom = len(m.cum)
}

func (m *prefixHeightModel) total() matrix.Float {
	m.ensureCum()
	return m.cum[len(m.cum)-1]
}

func (m *prefixHeightModel) offsetOf(index int) matrix.Float {
	m.ensureCum()
	index = min(max(index, 0), len(m.heights))
	return m.cum[index]
}

func (m *prefixHeightModel) heightOf(index int) matrix.Float {
	if index < 0 || index >= len(m.heights) {
		return 0
	}
	return m.heights[index]
}

func (m *prefixHeightModel) indexAt(y matrix.Float) int {
	if len(m.heights) == 0 {
		return 0
	}
	if y <= 0 {
		return 0
	}
	m.ensureCum()
	// Largest i such that cum[i] <= y. sort.Search finds the smallest i in
	// [0,n] with cum[i] > y; the row containing y is that minus one.
	n := len(m.heights)
	i := sort.Search(n+1, func(i int) bool { return m.cum[i] > y })
	i--
	return min(max(i, 0), n-1)
}

func (m *prefixHeightModel) setHeight(index int, h matrix.Float) {
	if index < 0 || index >= len(m.heights) {
		return
	}
	if h < 0 {
		h = 0
	}
	if m.heights[index] == h {
		return
	}
	m.heights[index] = h
	if index < m.dirtyFrom {
		m.dirtyFrom = index
	}
}

func (m *prefixHeightModel) reset() {
	for i := range m.heights {
		m.heights[i] = m.estimate
	}
	m.dirtyFrom = 0
}
