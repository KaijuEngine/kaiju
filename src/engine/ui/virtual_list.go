/******************************************************************************/
/* virtual_list.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

// VirtualList is a scrolling list that only materializes the items currently in
// view, recycling a small pool of cell elements as the user scrolls. That is
// what lets a list of 100k items cost the same to render as a list of 30: at any
// moment only a viewport-worth of cell elements are active.
//
// It is content-agnostic and not generic: the owner supplies a
// VirtualListDelegate that knows the item count and how to create/bind/unbind a
// cell element by index. Geometry — where each item sits and which are visible —
// is owned by a pluggable VirtualLayout strategy. The default strategy is a
// top-to-bottom vertical stack with item heights supplied either as a single
// fixed value (SetFixedRowHeight, the O(1) no-wrap path) or per-row
// (SetRowHeightFunc, the variable/wrapped path), so VirtualList behaves exactly
// as it always has unless SetLayout installs another strategy (horizontal, grid,
// masonry, ...). The rewritten TextArea is the first consumer (one row per text
// line); the kaiju-widgets RecyclerView is the second.
//
// All access is on the main thread, in the layout/update loop — no locks.
type VirtualList Panel

// VirtualAlign selects how ScrollToIndex positions a row within the viewport.
type VirtualAlign = int

const (
	// VirtualAlignStart aligns the row's top with the viewport top.
	VirtualAlignStart VirtualAlign = iota
	// VirtualAlignCenter centers the row vertically in the viewport.
	VirtualAlignCenter
	// VirtualAlignEnd aligns the row's bottom with the viewport bottom.
	VirtualAlignEnd
	// VirtualAlignNearest scrolls the minimum amount to make the row visible.
	VirtualAlignNearest
)

// VirtualListDelegate is the data source + row lifecycle for a VirtualList.
// Implementations own the backing data and answer purely by index.
type VirtualListDelegate interface {
	// RowCount is the total number of rows in the backing data.
	RowCount() int
	// CreateRow builds a fresh, reusable row element. It is called only when
	// the recycle pool is empty (at most a viewport-worth of times), and the
	// returned element is reused across many data indices via BindRow.
	CreateRow(man *Manager) *UI
	// BindRow populates an existing row element for the given data index.
	BindRow(index int, row *UI)
	// UnbindRow is called before a row leaves the viewport and returns to the
	// pool. Optional cleanup hook (clear text, hide children, etc.).
	UnbindRow(index int, row *UI)
}

const virtualListDefaultEstimate float32 = 20

// virtualListParkOffset is where recycled rows are moved (well above-left of any
// content) so they are scissor-clipped out of view while staying entity-active.
// We keep pooled rows active rather than SetActive(false): deactivating a row
// turns off its child glyph drawings, and reactivating + re-binding races the
// label render (it can mesh-then-deactivate and consume the render flag), leaving
// a permanent blank row. Parking sidesteps that entirely; the active pool stays
// bounded to a viewport's worth, so virtualization performance is unchanged.
// Content grows from origin (0,0) toward +x/+y, so a row parked at both
// {-1e6,-1e6} is off-screen for every scroll axis (vertical, horizontal, grid,
// masonry).
const virtualListParkOffset float32 = -1_000_000

// warmingRow is a pooled row that has been created but not yet shown. bornFrame
// records the frame it was created on; it becomes usable once the list has
// advanced past that frame (so it has been through a clean/render pass).
type warmingRow struct {
	row       *UI
	bornFrame uint64
}

type virtualListData struct {
	panelData
	content  *Panel
	delegate VirtualListDelegate
	// layout is the active geometry strategy (2-D). It defaults to linear (the
	// classic vertical stack), so existing consumers are unchanged.
	layout VirtualLayout
	// linear is the default vertical strategy, retained so the back-compat shims
	// (SetFixedRowHeight/SetRowHeightFunc/SetContentWidth and the vertical
	// accessors RowOffset/RowHeight/RowAt/TotalHeight) keep operating on it
	// regardless of which strategy is currently active.
	linear    *linearVerticalLayout
	overscan  int
	active    map[int]*UI
	free      []*UI
	warming   []warmingRow
	frame     uint64
	needsFill bool
	// pinned cells are skipped by the recycle/park loop: a drag-in-flight or
	// animating cell must outlive scrolling it out of the window.
	pinned map[*UI]struct{}
	// onAfterReflow, when set, runs at the END of reflow on the single-threaded
	// clean pass — the safe place for drag-follow and item-animation cell moves.
	onAfterReflow func()

	// reused scratch for the zero-allocation visible-vs-active diff.
	visibleSet map[int]struct{}
	visibleIdx []int

	lastWindowSize int

	// reflow change-gate: skip the body when nothing affecting geometry moved.
	lastScrollX   float32
	lastScrollY   float32
	lastViewportW float32
	lastViewportH float32
	lastTotal     float32
	haveLast      bool
}

func (d *virtualListData) innerPanelData() *panelData { return &d.panelData }

func (u *UI) ToVirtualList() *VirtualList      { return (*VirtualList)(u) }
func (vl *VirtualList) Base() *UI              { return (*UI)(vl) }
func (vl *VirtualList) Data() *virtualListData { return vl.elmData.(*virtualListData) }

func (vl *VirtualList) Init() {
	linear := newLinearVertical(virtualListDefaultEstimate)
	data := &virtualListData{
		active:     map[int]*UI{},
		overscan:   4,
		linear:     linear,
		layout:     linear,
		pinned:     map[*UI]struct{}{},
		visibleSet: map[int]struct{}{},
	}
	vl.elmData = data
	p := vl.Base().ToPanel()
	man := p.man.Value()
	// No background texture: the list is transparent so the owning element's
	// themed background shows through (a consumer that wants a panel background
	// paints it on a container behind the list).
	p.Init(nil, ElementTypeVirtualList)
	p.DontFitContent()
	p.SetOverflow(OverflowScroll)
	p.SetScrollDirection(PanelScrollDirectionVertical)

	data.content = man.Add().ToPanel()
	data.content.Init(nil, ElementTypePanel)
	data.content.DontFitContent()
	data.content.AllowClickThrough()
	p.AddChild(data.content.Base())
}

// Content returns the inner panel that holds the row elements. Consumers may
// parent overlays (cursors, selection rects) to it so they scroll with the rows.
func (vl *VirtualList) Content() *Panel { return vl.Data().content }

// RowOffset is the y of the top of row index in content space (vertical layout).
func (vl *VirtualList) RowOffset(index int) float32 { return vl.Data().linear.model.offsetOf(index) }

// RowHeight is the height of row index (vertical layout).
func (vl *VirtualList) RowHeight(index int) float32 { return vl.Data().linear.model.heightOf(index) }

// TotalHeight is the full scrollable content height (vertical layout).
func (vl *VirtualList) TotalHeight() float32 { return vl.Data().linear.model.total() }

// RowAt is the row index whose vertical span contains content-space y (vertical
// layout).
func (vl *VirtualList) RowAt(y float32) int { return vl.Data().linear.model.indexAt(y) }

// ViewportHeight is the visible content height of the list.
func (vl *VirtualList) ViewportHeight() float32 { return vl.viewportHeight() }

// SetContentWidth sets a minimum content width; the content panel is sized to
// max(viewportWidth, w). Used for horizontal scrolling of long, unwrapped rows
// in the default vertical layout. Pass 0 to track the viewport width (no
// horizontal scroll).
func (vl *VirtualList) SetContentWidth(w float32) {
	data := vl.Data()
	if data.linear.minContentW == w {
		return
	}
	data.linear.setMinContentW(w)
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

func (vl *VirtualList) SetDelegate(d VirtualListDelegate) {
	vl.Data().delegate = d
	vl.ReloadData()
}

func (vl *VirtualList) SetOverscan(rows int) {
	if rows < 0 {
		rows = 0
	}
	vl.Data().overscan = rows
	vl.invalidateWindow()
}

// SetFixedRowHeight switches the default vertical layout to fixed-height rows
// (every row is h tall). This is the O(1) path used for code (no soft wrap, one
// line per row).
func (vl *VirtualList) SetFixedRowHeight(h float32) {
	data := vl.Data()
	data.linear.setFixed(h)
	data.layout = data.linear
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// SetRowHeightFunc switches the default vertical layout to variable-height rows,
// measuring each row's height via fn. Used for wrapped / chat content.
func (vl *VirtualList) SetRowHeightFunc(fn func(index int) float32) {
	data := vl.Data()
	data.linear.setVariable(fn)
	data.layout = data.linear
	data.layout.SetCount(vl.rowCount())
	data.layout.Remeasure()
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// SetLayout installs a geometry strategy, switching the list among vertical
// (default), horizontal, grid, or masonry arrangements, and sets the panel scroll
// direction from the strategy's axis. Passing nil restores the default vertical
// strategy. The strategy's methods are pure math invoked only on the clean pass.
func (vl *VirtualList) SetLayout(l VirtualLayout) {
	data := vl.Data()
	if l == nil {
		l = data.linear
	}
	data.layout = l
	switch l.Axis() {
	case VirtualAxisHorizontal:
		(*Panel)(vl).SetScrollDirection(PanelScrollDirectionHorizontal)
	case VirtualAxisBoth:
		(*Panel)(vl).SetScrollDirection(PanelScrollDirectionBoth)
	default:
		(*Panel)(vl).SetScrollDirection(PanelScrollDirectionVertical)
	}
	l.SetCount(vl.rowCount())
	l.Remeasure()
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// Layout returns the active geometry strategy (so an owner can reach
// strategy-specific knobs it installed).
func (vl *VirtualList) Layout() VirtualLayout { return vl.Data().layout }

// OnAfterReflow registers a callback invoked at the end of every reflow, on the
// single-threaded clean pass — the only place it is safe to mutate cell layout in
// response to scrolling. Pass nil to clear. Used by the RecyclerView for
// drag-follow and item animations; the callback may re-dirty the list to keep
// animating and must stop doing so once it settles.
func (vl *VirtualList) OnAfterReflow(fn func()) { vl.Data().onAfterReflow = fn }

// PinRow marks a realized row as pinned: the recycle/park loop will not reclaim
// it when it scrolls out of the window (used to keep a drag-in-flight or
// animating cell alive). Unpin with UnpinRow.
func (vl *VirtualList) PinRow(row *UI) { vl.Data().pinned[row] = struct{}{} }

// UnpinRow removes a row's pin so it can be recycled normally again.
func (vl *VirtualList) UnpinRow(row *UI) { delete(vl.Data().pinned, row) }

// ReloadData re-reads the row count and rebinds the visible rows. Call after the
// backing data changes (count or content).
func (vl *VirtualList) ReloadData() {
	data := vl.Data()
	n := vl.rowCount()
	data.layout.SetCount(n)
	data.layout.Remeasure()
	vl.recycleAll()
	vl.clampScroll()
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// InvalidateRow re-measures a single row's extent (variable strategies) and
// reflows so the rows after it shift to their new positions.
func (vl *VirtualList) InvalidateRow(index int) {
	vl.Data().layout.Invalidate(index)
	vl.invalidateWindow()
	vl.reflow(true)
}

// RefreshVisible re-binds the rows currently on screen without changing the
// window (use when row content/styling changed but extents did not).
func (vl *VirtualList) RefreshVisible() {
	data := vl.Data()
	if data.delegate == nil {
		return
	}
	vp := vl.viewport()
	for idx, row := range data.active {
		rect := data.layout.RectOf(idx, vp)
		row.layout.SetOffset(rect.X, rect.Y)
		data.delegate.BindRow(idx, row)
	}
}

// VisibleRange returns the [first,last] data indices currently realized. If the
// list is empty last < first.
func (vl *VirtualList) VisibleRange() (first, last int) {
	data := vl.Data()
	first, last = -1, -1
	data.layout.VisibleAt(vl.scroll(), vl.viewport(), data.overscan, func(i int) {
		if first < 0 || i < first {
			first = i
		}
		if i > last {
			last = i
		}
	})
	if first < 0 {
		return 0, -1
	}
	return first, last
}

// ScrollToIndex scrolls so row index is visible according to align. It works for
// the active layout's primary axis (vertical or horizontal).
func (vl *VirtualList) ScrollToIndex(index int, align VirtualAlign) {
	data := vl.Data()
	n := vl.rowCount()
	if n == 0 {
		return
	}
	index = min(max(index, 0), n-1)
	p := (*Panel)(vl)
	vp := vl.viewport()
	rect := data.layout.RectOf(index, vp)
	horizontal := data.layout.Axis() == VirtualAxisHorizontal
	var top, size, vpMain, cur float32
	if horizontal {
		top, size, vpMain, cur = rect.X, rect.W, vp.X(), p.ScrollX()
	} else {
		top, size, vpMain, cur = rect.Y, rect.H, vp.Y(), p.ScrollY()
	}
	target := top
	switch align {
	case VirtualAlignStart:
		target = top
	case VirtualAlignEnd:
		target = top + size - vpMain
	case VirtualAlignCenter:
		target = top - (vpMain-size)*0.5
	case VirtualAlignNearest:
		if top < cur {
			target = top
		} else if top+size > cur+vpMain {
			target = top + size - vpMain
		} else {
			return
		}
	}
	if horizontal {
		p.SetScrollX(target)
	} else {
		p.SetScrollY(target)
	}
}

func (vl *VirtualList) onLayoutUpdating() {
	vl.reflow(false)
}

func (vl *VirtualList) update(deltaTime float64) {
	defer tracing.NewRegion("VirtualList.update").End()
	// NOTE: update() runs on a manager WORKER THREAD (updateFromManager), in
	// parallel with other elements. It must NOT create rows (man.Add is not
	// thread-safe) or render (the font cache is not concurrency-safe). All row
	// creation/binding/rendering is driven from onLayoutUpdating (the single-
	// threaded clean pass of this element's root). Here we only run the base
	// panel update (scroll handling) and ensure a reflow happens next clean if
	// the window is still being filled.
	vl.Base().ToPanel().update(deltaTime)
	if vl.Data().needsFill {
		vl.Base().SetDirty(DirtyTypeLayout)
	}
}

// --- internals ---------------------------------------------------------------

func (vl *VirtualList) rowCount() int {
	if d := vl.Data().delegate; d != nil {
		return d.RowCount()
	}
	return 0
}

func (vl *VirtualList) viewportHeight() float32 {
	return max(vl.layout.PixelSize().Y(), 0)
}

// viewport is the visible content size of the list (both axes, clamped to >= 0).
func (vl *VirtualList) viewport() matrix.Vec2 {
	ps := vl.layout.PixelSize()
	return matrix.Vec2{max(ps.X(), 0), max(ps.Y(), 0)}
}

// scroll is the current scroll offset in content space (both axes, positive).
func (vl *VirtualList) scroll() matrix.Vec2 {
	p := (*Panel)(vl)
	return matrix.Vec2{p.ScrollX(), p.ScrollY()}
}

func (vl *VirtualList) invalidateWindow() { vl.Data().haveLast = false }

func (vl *VirtualList) clampScroll() {
	p := (*Panel)(vl)
	p.SetScrollY(p.ScrollY())
}

func (vl *VirtualList) recycleAll() {
	data := vl.Data()
	for idx, row := range data.active {
		if _, pinned := data.pinned[row]; pinned {
			continue
		}
		if data.delegate != nil {
			data.delegate.UnbindRow(idx, row)
		}
		vl.parkRow(row)
		data.free = append(data.free, row)
		delete(data.active, idx)
	}
}

// parkRow moves a recycled/pooled row off-screen along BOTH axes so it is
// scissor-clipped out of view for any scroll axis while staying entity-active
// (see virtualListParkOffset).
func (vl *VirtualList) parkRow(row *UI) {
	row.layout.SetOffset(virtualListParkOffset, virtualListParkOffset)
}

// parkNewRow creates a new row element parked off-screen and queues it in the
// "warming" list tagged with the current frame. A row created via man.Add()
// during a frame is NOT in the manager's iteration list / clean tree for that
// frame, so it is only laid out/rendered the FOLLOWING frame. Using one before
// then renders it blank with no recovery, so a warming row is never handed to
// bind until a later frame (see promoteWarm). Creation only happens in reflow
// (once per frame, on the clean pass) so the frame tag is meaningful.
func (vl *VirtualList) parkNewRow() {
	data := vl.Data()
	row := data.delegate.CreateRow(vl.man.Value())
	row.entity.SetActive(true)
	data.content.AddChild(row)
	row.layout.SetPositioning(PositioningAbsolute)
	vl.parkRow(row)
	data.warming = append(data.warming, warmingRow{row: row, bornFrame: data.frame})
}

// promoteWarm moves rows that have been parked for at least one full frame from
// the warming queue into the usable free pool (they have now been through a
// clean/render pass and are safe to show). Called once per frame in reflow.
func (vl *VirtualList) promoteWarm() {
	data := vl.Data()
	if len(data.warming) == 0 {
		return
	}
	kept := data.warming[:0]
	for _, w := range data.warming {
		if w.bornFrame < data.frame {
			data.free = append(data.free, w.row)
		} else {
			kept = append(kept, w)
		}
	}
	data.warming = kept
}

// replenishPool grows the total pool (active + free + warming) up to target by
// creating parked rows. New rows go into the warming queue, never directly into
// free, so they are never shown the same frame they are created.
func (vl *VirtualList) replenishPool(target int) {
	data := vl.Data()
	for len(data.active)+len(data.free)+len(data.warming) < target {
		vl.parkNewRow()
	}
}

// virtualWindow is the pure visible-window computation for a 1-D vertical model:
// the [first,last] rows (inclusive) whose vertical span intersects
// [scrollY, scrollY+viewportH), expanded by overscan and clamped to [0,n).
// Returns last < first when empty. The default linear strategy uses it so the
// visible set stays identical to the pre-seam reflow.
func virtualWindow(model virtualHeightModel, n, overscan int, viewportH, scrollY float32) (first, last int) {
	if n == 0 {
		return 0, -1
	}
	first = max(model.indexAt(scrollY)-overscan, 0)
	last = min(model.indexAt(scrollY+viewportH)+overscan, n-1)
	return first, last
}

func (vl *VirtualList) reflow(force bool) {
	data := vl.Data()
	if data.delegate == nil {
		return
	}
	// reflow runs in the clean pass (onLayoutUpdating), single-threaded for this
	// element's root, so creating rows (man.Add) is safe here. Advance the warming
	// queue once per real frame: rows created during frame F are only promoted to
	// the usable pool once the host frame has moved past F, guaranteeing they have
	// been through a clean/render pass (they are not in the frame-F clean tree, so
	// using one in frame F would render it blank).
	if host := vl.man.Value().Host; host != nil {
		if f := host.Frame(); f != data.frame {
			data.frame = f
			vl.promoteWarm()
		}
	}

	p := (*Panel)(vl)
	n := vl.rowCount()
	data.layout.SetCount(n)
	vp := vl.viewport()
	scroll := matrix.Vec2{p.ScrollX(), p.ScrollY()}
	content := data.layout.ContentSize(vp)
	total := content.Y()
	if data.layout.Axis() == VirtualAxisHorizontal {
		total = content.X()
	}

	// Only do real work when the visible window can actually have changed — the
	// scroll offset, viewport, or content extent moved — UNLESS a previous pass
	// left a slot unfilled (needsFill), in which case keep running until the
	// window is complete (a fast scroll that outran the warm pool self-heals), or
	// a forced reflow (InvalidateRow) demands a rebind. Unlike the pre-seam gate
	// (which keyed on the contiguous [first,last] row window), this keys on
	// scroll/viewport/total so it works for any 2-D strategy. The post-reflow hook
	// still runs below even when gated, so animations keep ticking.
	gated := !force && !data.needsFill && data.haveLast &&
		approxEqf(scroll.X(), data.lastScrollX) && approxEqf(scroll.Y(), data.lastScrollY) &&
		approxEqf(vp.X(), data.lastViewportW) && approxEqf(vp.Y(), data.lastViewportH) &&
		approxEqf(total, data.lastTotal)

	if !gated {
		// Size the content panel to the full scroll extent so the panel's own
		// post-layout step computes both-axis maxScroll and the scrollbars
		// correctly.
		data.content.layout.Scale(max(content.X(), 1), max(content.Y(), 1))

		// Enumerate the visible set with zero allocation: clear the reused
		// membership set + index list, then collect via the strategy's visitor.
		clear(data.visibleSet)
		data.visibleIdx = data.visibleIdx[:0]
		data.layout.VisibleAt(scroll, vp, data.overscan, func(i int) {
			data.visibleSet[i] = struct{}{}
			data.visibleIdx = append(data.visibleIdx, i)
		})

		// Recycle rows that fell outside the visible set: unbind and park them
		// off-screen (kept entity-active). Pinned rows (drag/animation) are kept.
		for idx, row := range data.active {
			if _, vis := data.visibleSet[idx]; vis {
				continue
			}
			if _, pinned := data.pinned[row]; pinned {
				continue
			}
			data.delegate.UnbindRow(idx, row)
			vl.parkRow(row)
			data.free = append(data.free, row)
			delete(data.active, idx)
		}

		// Bind every visible slot, but ONLY with a "warm" row — one that already
		// existed (and was laid out/rendered) before this pass. A row created this
		// frame is NOT in the manager's iteration list or this frame's clean tree,
		// so showing it immediately renders it blank with no recovery. The recycled
		// rows just parked are warm; so are the rows in the free pool.
		warm := len(data.free)
		data.needsFill = false
		for _, i := range data.visibleIdx {
			if row, ok := data.active[i]; ok {
				if force {
					rect := data.layout.RectOf(i, vp)
					row.layout.SetOffset(rect.X, rect.Y)
					data.delegate.BindRow(i, row)
				}
				continue
			}
			if warm <= 0 {
				// No warm row available for this slot this frame. Leave it empty,
				// flag the window incomplete; update() grows the pool and needsFill
				// keeps reflow running until it can be filled. We never
				// create-and-show a row here (that renders blank with no recovery).
				data.needsFill = true
				continue
			}
			row := data.free[0]
			data.free = data.free[1:]
			warm--
			data.active[i] = row
			rect := data.layout.RectOf(i, vp)
			row.layout.SetPositioning(PositioningAbsolute)
			row.layout.SetOffset(rect.X, rect.Y)
			data.delegate.BindRow(i, row)
			// Mark the row dirty so this clean pass (we are in onLayoutUpdating)
			// re-runs its layout + render at the new offset; otherwise a row laid
			// out while parked stays at the parked position.
			row.SetDirty(DirtyTypeGenerated)
		}

		data.lastScrollX, data.lastScrollY = scroll.X(), scroll.Y()
		data.lastViewportW, data.lastViewportH = vp.X(), vp.Y()
		data.lastTotal = total
		data.lastWindowSize = len(data.visibleIdx)
		data.haveLast = true

		// Top the pool up to a warm buffer big enough to fill the whole visible
		// window in one go after rows warm. New rows go into the warming queue
		// (used a later frame). Done here in the single-threaded clean pass so
		// man.Add is safe.
		vl.replenishPool(data.lastWindowSize + data.overscan*2 + 8)

		// If the window could not be fully filled this frame (warm pool too small),
		// make sure another clean pass runs so it gets filled as the pool warms up.
		if data.needsFill {
			vl.Base().SetDirty(DirtyTypeLayout)
		}
	}

	// Run the post-reflow hook on the clean pass (drag-follow, animation lerp). It
	// runs even when the recycle/bind work was gated, so animations keep ticking
	// while the list is otherwise idle.
	if data.onAfterReflow != nil {
		data.onAfterReflow()
	}
}

func approxEqf(a, b float32) bool {
	d := a - b
	return d < 0.01 && d > -0.01
}
