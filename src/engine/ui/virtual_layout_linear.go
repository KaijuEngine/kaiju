/******************************************************************************/
/* virtual_layout_linear.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "kaijuengine.com/matrix"

// linearVerticalLayout is the default VirtualLayout: a top-to-bottom vertical
// stack, one item per main-axis line, exactly the behavior VirtualList had
// before the layout seam existed. It is a thin wrapper over the existing
// virtualHeightModel implementations (fixedHeightModel for the O(1) no-wrap path,
// prefixHeightModel for the variable/wrapped path), so TextArea — which only ever
// uses the default — sees byte-identical geometry.
//
// RectOf always returns X==0 and W==content width, so the VirtualList positions
// every cell at x=0 just as it did when reflow hard-coded SetOffset(0, y). The
// minContentW field carries SetContentWidth's value (horizontal scroll of long,
// unwrapped rows) into ContentSize/RectOf.
type linearVerticalLayout struct {
	fixed       *fixedHeightModel
	prefix      *prefixHeightModel
	model       virtualHeightModel // == fixed or prefix
	heightFn    func(index int) matrix.Float
	minContentW matrix.Float
}

func newLinearVertical(estimate matrix.Float) *linearVerticalLayout {
	f := newFixedHeightModel(estimate)
	return &linearVerticalLayout{fixed: f, model: f}
}

// --- VirtualLayout -----------------------------------------------------------

func (l *linearVerticalLayout) Axis() VirtualAxis { return VirtualAxisVertical }

func (l *linearVerticalLayout) SetCount(n int) { l.model.setCount(n) }

func (l *linearVerticalLayout) Count() int { return l.model.count() }

func (l *linearVerticalLayout) ContentSize(viewport matrix.Vec2) matrix.Vec2 {
	return matrix.Vec2{l.contentWidth(viewport), l.model.total()}
}

func (l *linearVerticalLayout) VisibleAt(scroll, viewport matrix.Vec2, overscan int, visit VirtualVisitor) {
	// Reuse the exact, already-tested window computation so the visible set is
	// identical to the pre-seam reflow.
	first, last := virtualWindow(l.model, l.model.count(), overscan, viewport.Y(), scroll.Y())
	for i := first; i <= last; i++ {
		visit(i)
	}
}

func (l *linearVerticalLayout) RectOf(index int, viewport matrix.Vec2) VirtualRect {
	return VirtualRect{
		X: 0,
		Y: l.model.offsetOf(index),
		W: l.contentWidth(viewport),
		H: l.model.heightOf(index),
	}
}

func (l *linearVerticalLayout) Invalidate(index int) {
	// Fixed model ignores setHeight; variable model re-pulls the row's height.
	if l.heightFn != nil {
		l.model.setHeight(index, l.heightFn(index))
	}
}

func (l *linearVerticalLayout) Remeasure() {
	// Only the variable path measures rows; fixed rows share one height.
	if l.model != l.prefix || l.heightFn == nil {
		return
	}
	l.prefix.reset()
	for i := range l.prefix.count() {
		l.prefix.setHeight(i, l.heightFn(i))
	}
}

// --- back-compat configuration (driven by the VirtualList shims) -------------

// setFixed switches to fixed-height rows (the O(1) path).
func (l *linearVerticalLayout) setFixed(h matrix.Float) {
	l.fixed.setRowHeight(h)
	l.model = l.fixed
	l.heightFn = nil
}

// setVariable switches to per-row measured heights via fn.
func (l *linearVerticalLayout) setVariable(fn func(index int) matrix.Float) {
	if l.prefix == nil {
		l.prefix = newPrefixHeightModel(virtualListDefaultEstimate)
	}
	l.model = l.prefix
	l.heightFn = fn
}

// setMinContentW records SetContentWidth's minimum content width.
func (l *linearVerticalLayout) setMinContentW(w matrix.Float) { l.minContentW = w }

// contentWidth is the content panel width: the viewport width, widened to
// minContentW when a larger one was requested (horizontal scroll of long rows).
func (l *linearVerticalLayout) contentWidth(viewport matrix.Vec2) matrix.Float {
	w := max(viewport.X(), 1)
	if l.minContentW > w {
		w = l.minContentW
	}
	return w
}
