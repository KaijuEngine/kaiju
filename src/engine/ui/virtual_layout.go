/******************************************************************************/
/* virtual_layout.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "kaijuengine.com/matrix"

// VirtualLayout is the pluggable geometry strategy for a VirtualList. It owns
// ALL item geometry in 2-D content space; the VirtualList core keeps doing the
// hard parts (recycle pool, warming/park, the visible-vs-active diff, threading)
// and only asks the strategy pure questions: how big is the content, which items
// are visible, and where each one sits.
//
// Because every method is pure math over (scroll, viewport, index) — it must NOT
// create elements (man.Add), render, touch fonts, or reach the Host — a
// VirtualLayout can be implemented from ANY package. The engine ships exactly
// one implementation, linearVerticalLayout, installed by default so a VirtualList
// behaves identically to before unless SetLayout installs another strategy. The
// kaiju-widgets RecyclerView supplies horizontal, grid, and masonry strategies;
// third parties can supply their own.
//
// All methods are invoked on the VirtualList's single-threaded clean pass
// (onLayoutUpdating -> reflow), the same place row creation happens, so a
// strategy never has to think about concurrency.
type VirtualLayout interface {
	// Axis is the primary scroll axis the strategy drives. The VirtualList sets
	// its panel scroll direction to match when the strategy is installed.
	Axis() VirtualAxis
	// SetCount sets the item count. It is cheap (O(1)); strategies that need an
	// O(n) build (e.g. masonry column assignment) defer it to the first geometry
	// query, since the build depends on the viewport.
	SetCount(n int)
	// Count is the current item count.
	Count() int
	// ContentSize is the full scrollable content size for the given viewport
	// (the viewport drives column/row counts for grid/masonry). The VirtualList
	// feeds it to the content panel so both-axis max-scroll and scrollbars are
	// correct.
	ContentSize(viewport matrix.Vec2) matrix.Vec2
	// VisibleAt enumerates, via visit, every item index whose rect intersects the
	// window [scroll, scroll+viewport) expanded by overscan items along the main
	// axis. It MUST NOT allocate. The VirtualList uses it to diff which cells are
	// on screen.
	VisibleAt(scroll, viewport matrix.Vec2, overscan int, visit VirtualVisitor)
	// RectOf is the content-space rect (position + size) of an item for the given
	// viewport. The VirtualList positions the recycled cell at (X,Y); the cell's
	// size is applied by the owner's BindRow (the engine never Scales cells, so
	// content that self-sizes — like TextArea's labels — is undisturbed).
	RectOf(index int, viewport matrix.Vec2) VirtualRect
	// Invalidate re-derives geometry for a single item whose size source may have
	// changed (a measured/variable extent), shifting the items after it. Fixed
	// strategies treat it as a no-op.
	Invalidate(index int)
	// Remeasure re-derives ALL item geometry from the strategy's size source
	// (O(n), or deferred to the next geometry query). The VirtualList calls it
	// when the backing data is reloaded. Fixed strategies treat it as a no-op.
	Remeasure()
}

// VirtualAxis is the primary scroll axis a VirtualLayout drives. Vertical is the
// default so existing VirtualList consumers are unaffected.
type VirtualAxis = int32

const (
	// VirtualAxisVertical scrolls along Y (the default).
	VirtualAxisVertical VirtualAxis = iota
	// VirtualAxisHorizontal scrolls along X.
	VirtualAxisHorizontal
	// VirtualAxisBoth scrolls along both axes (grids/masonry that can overflow
	// the viewport in either direction).
	VirtualAxisBoth
)

// VirtualRect is the content-space rectangle of a virtualized item: position
// (X,Y) of its top-left corner and size (W,H), measured in the same content
// space the VirtualList's inner content panel lives in (origin top-left, +x
// right, +y down). It is a pure data carrier (no behavior), an exported struct
// on a hot path like ColorSpan, so strategies return it by value with no
// allocation.
type VirtualRect struct {
	X, Y, W, H float32
}

// VirtualVisitor receives each visible item index during a VisibleAt
// enumeration. It exists so the window can be walked with zero allocation (no
// returned slice).
type VirtualVisitor func(index int)
