/******************************************************************************/
/* virtual_list.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"kaijuengine.com/platform/profiler/tracing"
)

// VirtualList is a vertically-scrolling list that only materializes the rows
// currently in view, recycling a small pool of row elements as the user
// scrolls. That is what lets a list of 100k rows cost the same to render as a
// list of 30: at any moment only a viewport-worth of row elements are active.
//
// It is content-agnostic and not generic: the owner supplies a
// VirtualListDelegate that knows the row count and how to create/bind/unbind a
// row element by index, and supplies row heights either as a single fixed value
// (SetFixedRowHeight, the O(1) no-wrap path) or per-row (SetRowHeightFunc, the
// variable/wrapped path). The rewritten TextArea is the first consumer (one row
// per text line); Select's option list is a natural follow-up.
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

// virtualListParkOffset is where recycled rows are moved (well above any content)
// so they are scissor-clipped out of view while staying entity-active. We keep
// pooled rows active rather than SetActive(false): deactivating a row turns off
// its child glyph drawings, and reactivating + re-binding races the label render
// (it can mesh-then-deactivate and consume the render flag), leaving a permanent
// blank row. Parking sidesteps that entirely; the active pool stays bounded to a
// viewport's worth, so virtualization performance is unchanged.
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
	content        *Panel
	delegate       VirtualListDelegate
	model          virtualHeightModel
	fixedModel     *fixedHeightModel
	prefixModel    *prefixHeightModel
	heightFn       func(index int) float32
	overscan       int
	minContentW    float32
	active         map[int]*UI
	free           []*UI
	warming        []warmingRow
	frame          uint64
	needsFill      bool
	lastWindowSize int

	lastFirst     int
	lastLast      int
	lastScrollY   float32
	lastViewportH float32
	lastTotal     float32
	haveLast      bool
}

func (d *virtualListData) innerPanelData() *panelData { return &d.panelData }

func (u *UI) ToVirtualList() *VirtualList      { return (*VirtualList)(u) }
func (vl *VirtualList) Base() *UI              { return (*UI)(vl) }
func (vl *VirtualList) Data() *virtualListData { return vl.elmData.(*virtualListData) }

func (vl *VirtualList) Init() {
	data := &virtualListData{
		active:     map[int]*UI{},
		overscan:   4,
		fixedModel: newFixedHeightModel(virtualListDefaultEstimate),
		lastLast:   -1,
	}
	data.model = data.fixedModel
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

// RowOffset is the y of the top of row index in content space.
func (vl *VirtualList) RowOffset(index int) float32 { return vl.Data().model.offsetOf(index) }

// RowHeight is the height of row index.
func (vl *VirtualList) RowHeight(index int) float32 { return vl.Data().model.heightOf(index) }

// TotalHeight is the full scrollable content height.
func (vl *VirtualList) TotalHeight() float32 { return vl.Data().model.total() }

// RowAt is the row index whose vertical span contains content-space y.
func (vl *VirtualList) RowAt(y float32) int { return vl.Data().model.indexAt(y) }

// ViewportHeight is the visible content height of the list.
func (vl *VirtualList) ViewportHeight() float32 { return vl.viewportHeight() }

// SetContentWidth sets a minimum content width; the content panel is sized to
// max(viewportWidth, w). Used for horizontal scrolling of long, unwrapped rows.
// Pass 0 to track the viewport width (no horizontal scroll).
func (vl *VirtualList) SetContentWidth(w float32) {
	data := vl.Data()
	if data.minContentW == w {
		return
	}
	data.minContentW = w
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

// SetFixedRowHeight switches the list to fixed-height rows (every row is h tall).
// This is the O(1) path used for code (no soft wrap, one line per row).
func (vl *VirtualList) SetFixedRowHeight(h float32) {
	data := vl.Data()
	if data.model == data.fixedModel && approxEqf(data.fixedModel.rowHeight, h) {
		return
	}
	data.fixedModel.setRowHeight(h)
	data.model = data.fixedModel
	data.heightFn = nil
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// SetRowHeightFunc switches the list to variable-height rows, measuring each
// row's height via fn. Used for wrapped / chat content.
func (vl *VirtualList) SetRowHeightFunc(fn func(index int) float32) {
	data := vl.Data()
	if data.prefixModel == nil {
		data.prefixModel = newPrefixHeightModel(virtualListDefaultEstimate)
	}
	data.model = data.prefixModel
	data.heightFn = fn
	vl.remeasureAll()
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// ReloadData re-reads the row count and rebinds the visible rows. Call after the
// backing data changes (count or content).
func (vl *VirtualList) ReloadData() {
	data := vl.Data()
	n := vl.rowCount()
	data.model.setCount(n)
	vl.remeasureAll()
	vl.recycleAll()
	vl.clampScroll()
	vl.invalidateWindow()
	vl.Base().SetDirty(DirtyTypeLayout)
}

// InvalidateRow re-measures a single row's height (variable mode) and reflows so
// the rows below it shift to their new positions.
func (vl *VirtualList) InvalidateRow(index int) {
	data := vl.Data()
	if data.heightFn != nil {
		data.model.setHeight(index, data.heightFn(index))
	}
	vl.invalidateWindow()
	vl.reflow(true)
}

// RefreshVisible re-binds the rows currently on screen without changing the
// window (use when row content/styling changed but heights did not).
func (vl *VirtualList) RefreshVisible() {
	data := vl.Data()
	if data.delegate == nil {
		return
	}
	for idx, row := range data.active {
		row.layout.SetOffset(0, data.model.offsetOf(idx))
		data.delegate.BindRow(idx, row)
	}
}

// VisibleRange returns the [first,last] data indices currently realized. If the
// list is empty last < first.
func (vl *VirtualList) VisibleRange() (first, last int) {
	return vl.windowFor(vl.viewportHeight(), (*Panel)(vl).ScrollY())
}

// ScrollToIndex scrolls so row index is visible according to align.
func (vl *VirtualList) ScrollToIndex(index int, align VirtualAlign) {
	data := vl.Data()
	n := vl.rowCount()
	if n == 0 {
		return
	}
	index = min(max(index, 0), n-1)
	top := data.model.offsetOf(index)
	h := data.model.heightOf(index)
	vh := vl.viewportHeight()
	target := top
	switch align {
	case VirtualAlignStart:
		target = top
	case VirtualAlignEnd:
		target = top + h - vh
	case VirtualAlignCenter:
		target = top - (vh-h)*0.5
	case VirtualAlignNearest:
		scrollY := (*Panel)(vl).ScrollY()
		if top < scrollY {
			target = top
		} else if top+h > scrollY+vh {
			target = top + h - vh
		} else {
			return
		}
	}
	(*Panel)(vl).SetScrollY(target)
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

func (vl *VirtualList) invalidateWindow() { vl.Data().haveLast = false }

func (vl *VirtualList) clampScroll() {
	p := (*Panel)(vl)
	p.SetScrollY(p.ScrollY())
}

// remeasureAll fills in every row height for the variable-height model up front.
// For fixed-height this is a no-op. The huge-list requirement is the fixed path,
// so paying O(n) here for the variable (chat) path is acceptable.
func (vl *VirtualList) remeasureAll() {
	data := vl.Data()
	if data.model != data.prefixModel || data.heightFn == nil {
		return
	}
	data.prefixModel.reset()
	n := vl.rowCount()
	for i := range n {
		data.prefixModel.setHeight(i, data.heightFn(i))
	}
}

func (vl *VirtualList) recycleAll() {
	data := vl.Data()
	for idx, row := range data.active {
		if data.delegate != nil {
			data.delegate.UnbindRow(idx, row)
		}
		row.layout.SetOffset(0, virtualListParkOffset)
		data.free = append(data.free, row)
		delete(data.active, idx)
	}
}

// parkNewRow creates a new row element parked off-screen and queues it in the
// "warming" list tagged with the current frame. A row created via man.Add()
// during a frame is NOT in the manager's iteration list / clean tree for that
// frame, so it is only laid out/rendered the FOLLOWING frame. Using one before
// then renders it blank with no recovery, so a warming row is never handed to
// bind until a later frame (see promoteWarm). Creation only happens in update()
// (once per frame) so the frame tag is meaningful.
func (vl *VirtualList) parkNewRow() {
	data := vl.Data()
	row := data.delegate.CreateRow(vl.man.Value())
	row.entity.SetActive(true)
	data.content.AddChild(row)
	row.layout.SetPositioning(PositioningAbsolute)
	row.layout.SetOffset(0, virtualListParkOffset)
	data.warming = append(data.warming, warmingRow{row: row, bornFrame: data.frame})
}

// promoteWarm moves rows that have been parked for at least one full frame from
// the warming queue into the usable free pool (they have now been through a
// clean/render pass and are safe to show). Called once per frame in update().
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

// windowFor computes the visible [first,last] index range for a viewport height
// and scroll offset, expanded by overscan and clamped to the data range.
func (vl *VirtualList) windowFor(viewportH, scrollY float32) (first, last int) {
	data := vl.Data()
	return virtualWindow(data.model, vl.rowCount(), data.overscan, viewportH, scrollY)
}

// virtualWindow is the pure visible-window computation: the [first,last] rows
// (inclusive) whose vertical span intersects [scrollY, scrollY+viewportH),
// expanded by overscan and clamped to [0,n). Returns last < first when empty.
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
	data.model.setCount(n)
	ps := vl.layout.PixelSize()
	viewportH := max(ps.Y(), 0)
	width := max(max(ps.X(), 1), data.minContentW)
	total := data.model.total()
	scrollY := p.ScrollY()
	first, last := vl.windowFor(viewportH, scrollY)

	// Only do work when the visible WINDOW (row range), viewport, or content
	// actually changes — NOT on every scroll pixel — UNLESS a previous pass left
	// some visible slot unfilled (needsFill), in which case keep running until the
	// window is complete (so a fast scroll that outran the warm pool self-heals).
	if !force && !data.needsFill && data.haveLast &&
		first == data.lastFirst && last == data.lastLast &&
		approxEqf(viewportH, data.lastViewportH) &&
		approxEqf(total, data.lastTotal) {
		return
	}

	// Size the content panel to the full scroll height so the panel's own
	// post-layout step computes maxScroll and the scrollbar correctly.
	data.content.layout.Scale(width, max(total, 1))

	// Recycle rows that fell outside the window: unbind and park them off-screen
	// (kept entity-active — see virtualListParkOffset). Parked rows go back into
	// the free pool and, being already laid-out/rendered, are "warm".
	for idx, row := range data.active {
		if idx < first || idx > last {
			data.delegate.UnbindRow(idx, row)
			row.layout.SetOffset(0, virtualListParkOffset)
			data.free = append(data.free, row)
			delete(data.active, idx)
		}
	}
	// Bind every visible slot, but ONLY ever with a "warm" row — one that already
	// existed (and was laid out/rendered) before this pass. A row created this
	// frame is NOT in the manager's iteration list or this frame's clean tree, so
	// if it were shown immediately it would render blank with no way to recover.
	// All rows in the free pool at this point are warm (parked >= 1 frame ago);
	// the recycled rows just added are also warm. We consume only that many.
	warm := len(data.free)
	data.needsFill = false
	for i := first; i <= last; i++ {
		if _, ok := data.active[i]; ok {
			if force {
				row := data.active[i]
				row.layout.SetOffset(0, data.model.offsetOf(i))
				data.delegate.BindRow(i, row)
			}
			continue
		}
		if warm <= 0 {
			// No warm row available for this slot this frame. Leave it empty for
			// now and flag that the window is incomplete; update() grows the pool
			// and needsFill keeps reflow running until it can be filled. We never
			// create-and-show a row here (that renders blank with no recovery).
			data.needsFill = true
			continue
		}
		row := data.free[0]
		data.free = data.free[1:]
		warm--
		data.active[i] = row
		row.layout.SetPositioning(PositioningAbsolute)
		row.layout.SetOffset(0, data.model.offsetOf(i))
		data.delegate.BindRow(i, row)
		// Mark the row dirty so this clean pass (we are in onLayoutUpdating) re-runs
		// its layout + render at the new offset; otherwise a row that was last laid
		// out while parked stays at the parked position.
		row.SetDirty(DirtyTypeGenerated)
	}

	data.lastFirst, data.lastLast = first, last
	data.lastViewportH, data.lastTotal = viewportH, total
	data.lastWindowSize = last - first + 1
	data.haveLast = true

	// Top the pool up to a warm buffer big enough to fill the whole visible window
	// in one go after rows warm. New rows go into the warming queue (used a later
	// frame). Done here in the single-threaded clean pass so man.Add is safe.
	vl.replenishPool(data.lastWindowSize + data.overscan*2 + 8)

	// If the window could not be fully filled this frame (warm pool too small),
	// make sure another clean pass runs so it gets filled as the pool warms up.
	if data.needsFill {
		vl.Base().SetDirty(DirtyTypeLayout)
	}
}

func approxEqf(a, b float32) bool {
	d := a - b
	return d < 0.01 && d > -0.01
}
