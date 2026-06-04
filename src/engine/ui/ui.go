/******************************************************************************/
/* ui.go                                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"log/slog"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"weak"

	"kaijuengine.com/build"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
	"kaijuengine.com/rendering"
)

// uiDiagEnabled gates the temporary layout-convergence diagnostics. Enable with
// KAIJU_UI_DIAG=1 to log slow / non-converging UI.Clean passes and the elements
// that stay dirty. TEMP: remove once the layout perf issue is resolved.
var uiDiagEnabled = os.Getenv("KAIJU_UI_DIAG") != "" && build.Debug
var uiDiagCount atomic.Int64

// diagCapturing/diagDirtyCounts capture WHICH call site re-dirties each element
// during the one instrumented settle-pass in uiCleanDiag. Single-root use only
// (the gallery has one UI root) so the plain map needs no synchronization.
var diagCapturing atomic.Bool
var diagDirtyCounts map[string]int

// diagDirtySource walks the stack past the dirty plumbing (SetDirty/
// layoutChanged/setDirtyInternal) to the real setter that requested the dirty,
// returning "shortFunc:line". TEMP diagnostic (uiDiagEnabled).
func diagDirtySource() string {
	var pcs [16]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	for {
		f, more := frames.Next()
		nm := f.Function
		if strings.Contains(nm, "layoutChanged") || strings.HasSuffix(nm, ".SetDirty") ||
			strings.Contains(nm, "setDirtyInternal") || strings.Contains(nm, "diagDirtySource") {
			if !more {
				break
			}
			continue
		}
		if idx := strings.LastIndexByte(nm, '/'); idx >= 0 {
			nm = nm[idx+1:]
		}
		return nm + ":" + strconv.Itoa(f.Line)
	}
	return "?"
}

func diagRecord(ui *UI, dirtyType DirtyType) {
	if diagDirtyCounts == nil {
		return
	}
	diagDirtyCounts[diagAncestry(ui)+" elm="+strconv.Itoa(int(ui.elmType))+
		" dirty="+strconv.Itoa(dirtyType)+" <- "+diagDirtySource()]++
}

// diagAncestry returns the chain of up to 6 ancestor entity names (· for an
// unnamed entity), nearest first, to locate otherwise-unnamed elements in the
// tree. TEMP diagnostic (uiDiagEnabled).
func diagAncestry(t *UI) string {
	var b strings.Builder
	e := &t.entity
	for i := 0; i < 6 && e != nil; i++ {
		if i > 0 {
			b.WriteByte('<')
		}
		if nm := e.Name(); nm != "" {
			b.WriteString(nm)
		} else {
			b.WriteByte('*') // unnamed entity
		}
		e = e.Parent
	}
	return b.String()
}

type DirtyType = int
type ElementType = uint8
type uiBits uint16

const (
	DirtyTypeNone = iota
	DirtyTypeLayout
	DirtyTypeResize
	DirtyTypeGenerated
	DirtyTypeColorChange
	DirtyTypeScissor
	DirtyTypeParent
	DirtyTypeParentLayout
	DirtyTypeParentResize
	DirtyTypeParentGenerated
	DirtyTypeParentReGenerated
	DirtyTypeParentColorChange
	DirtyTypeParentScissor
)

const (
	ElementTypeLabel = ElementType(iota)
	ElementTypePanel
	ElementTypeButton
	ElementTypeCheckbox
	ElementTypeImage
	ElementTypeInput
	ElementTypeProgressBar
	ElementTypeSelect
	ElementTypeSlider
	ElementTypeTextArea
)

const (
	uiBitsIsScrolling uiBits = 1 << iota
	uiBitsHovering
	uiBitsCantMiss
	uiBitsIsDown
	uiBitsIsRightDown
	uiBitsDrag
	uiBitsLastActive
	uiBitsDontClean
	uiBitsDisabled
)

type UIElementData interface {
	innerPanelData() *panelData
}

type UI struct {
	man              weak.Pointer[Manager]
	entity           engine.Entity
	elmData          UIElementData
	events           [EventTypeEnd]events.Event
	postLayoutUpdate func()
	render           func()
	layout           Layout
	dragStartPos     matrix.Vec3
	downPos          matrix.Vec2
	elmType          ElementType
	dirtyType        DirtyType
	shaderData       *ShaderData
	textureSize      matrix.Vec2
	lastClick        float64
	poolId           pooling.PoolGroupId
	id               pooling.PoolIndex
	flags            uiBits
}

func (b uiBits) hovering() bool     { return b&uiBitsHovering != 0 }
func (b uiBits) cantMiss() bool     { return b&uiBitsCantMiss != 0 }
func (b uiBits) isDown() bool       { return b&uiBitsIsDown != 0 }
func (b uiBits) isRightDown() bool  { return b&uiBitsIsRightDown != 0 }
func (b uiBits) drag() bool         { return b&uiBitsDrag != 0 }
func (b uiBits) lastActive() bool   { return b&uiBitsLastActive != 0 }
func (b uiBits) dontClean() bool    { return b&uiBitsDontClean != 0 }
func (b uiBits) disabled() bool     { return b&uiBitsDisabled != 0 }
func (b *uiBits) setHovering()      { *b |= uiBitsHovering }
func (b *uiBits) setCantMiss()      { *b |= uiBitsCantMiss }
func (b *uiBits) setIsDown()        { *b |= uiBitsIsDown }
func (b *uiBits) setIsRightDown()   { *b |= uiBitsIsRightDown }
func (b *uiBits) setDrag()          { *b |= uiBitsDrag }
func (b *uiBits) setLastActive()    { *b |= uiBitsLastActive }
func (b *uiBits) setDontClean()     { *b |= uiBitsDontClean }
func (b *uiBits) setDisabled()      { *b |= uiBitsDisabled }
func (b *uiBits) resetHovering()    { *b &= ^uiBitsHovering }
func (b *uiBits) resetCantMiss()    { *b &= ^uiBitsCantMiss }
func (b *uiBits) resetIsDown()      { *b &= ^uiBitsIsDown }
func (b *uiBits) resetIsRightDown() { *b &= ^uiBitsIsRightDown }
func (b *uiBits) resetDrag()        { *b &= ^uiBitsDrag }
func (b *uiBits) resetLastActive()  { *b &= ^uiBitsLastActive }
func (b *uiBits) resetDontClean()   { *b &= ^uiBitsDontClean }
func (b *uiBits) resetDisabled()    { *b &= ^uiBitsDisabled }

func (ui *UI) IsActive() bool { return ui.entity.IsActive() }
func (ui *UI) IsDown() bool   { return ui.flags.isDown() }
func (ui *UI) IsValid() bool  { return ui.elmData != nil }

func (ui *UI) init(textureSize matrix.Vec2) {
	defer tracing.NewRegion("UI.init").End()
	if ui.postLayoutUpdate == nil {
		ui.postLayoutUpdate = func() {}
	}
	if ui.render == nil {
		ui.render = func() { ui.events[EventTypeRender].Execute() }
	}
	ui.shaderData = &ShaderData{
		ShaderDataBase: rendering.NewShaderDataBase(),
	}
	ui.shaderData.Scissor = matrix.Vec4{-matrix.FloatMax, -matrix.FloatMax, matrix.FloatMax, matrix.FloatMax}
	ui.entity.AddNamedData(EntityDataName, ui)
	ui.textureSize = textureSize
	ui.layout.initialize(ui)
	host := ui.man.Value().Host
	rzId := host.Window.OnResize.Add(func() {
		ui.SetDirty(DirtyTypeResize)
		if ui.Type() == ElementTypeInput {
			// Labels that make up the input box don't always re-render with
			// minor events, this is a full window resize so it needs to happen.
			ui.ToInput().forceLabelAndPlaceholderRerender()
		} else if ui.Type() == ElementTypeTextArea {
			ui.ToTextArea().forceLabelAndPlaceholderRerender()
		}
	})
	ui.entity.OnDestroy.Add(func() {
		host.Window.OnResize.Remove(rzId)
		ui.shaderData.Destroy()
		ui.events[EventTypeDestroy].Execute()
		ui.elmData = nil
		ui.postLayoutUpdate = nil
		ui.render = nil
		ui.layout.ui = nil
		ui.layout.Stylizer = nil
		for i := range ui.events {
			ui.events[i].Clear()
		}
		if ui.man.Value() != nil {
			ui.man.Value().Remove(ui)
		}
	})
}

func (ui *UI) Entity() *engine.Entity          { return &ui.entity }
func (ui *UI) Layout() *Layout                 { return &ui.layout }
func (ui *UI) hasScissor() bool                { return ui.shaderData.Scissor.X() > -matrix.FloatMax }
func (ui *UI) selfScissor() matrix.Vec4        { return ui.shaderData.Scissor }
func (ui *UI) dirty() DirtyType                { return ui.dirtyType }
func (ui *UI) ShaderData() *ShaderData         { return ui.shaderData }
func (ui *UI) IsType(elmType ElementType) bool { return ui.elmType == elmType }
func (ui *UI) Type() ElementType               { return ui.elmType }

func (ui *UI) Host() *engine.Host {
	if ui.man.Value() != nil {
		return ui.man.Value().Host
	}
	return nil
}

func (ui *UI) SetDontClean(val bool) {
	if val {
		ui.flags.setDontClean()
	} else {
		ui.flags.resetDontClean()
	}
}

func (ui *UI) IsDisabled() bool {
	return ui.flags.disabled()
}

func (ui *UI) SetDisabled(disabled bool) {
	if ui.IsDisabled() == disabled {
		return
	}
	if disabled {
		ui.flags.setDisabled()
		ui.flags.resetIsDown()
		ui.flags.resetIsRightDown()
		ui.flags.resetDrag()
		ui.flags.resetCantMiss()
		ui.flags.resetHovering()
		if man := ui.man.Value(); man != nil && man.Host != nil && man.Host.Window != nil {
			man.Host.Window.CursorStandard()
		}
		switch ui.Type() {
		case ElementTypeInput:
			ui.ToInput().removeFocusWithoutEvents()
		case ElementTypeTextArea:
			ui.ToTextArea().removeFocusWithoutEvents()
		case ElementTypeSelect:
			ui.ToSelect().collapse()
		}
	} else {
		ui.flags.resetDisabled()
	}
	if ui.IsValid() {
		ui.SetDirty(DirtyTypeGenerated)
	}
}

func (ui *UI) disabledBlocksEvent(evtType EventType) bool {
	if !ui.IsDisabled() {
		return false
	}
	switch evtType {
	case EventTypeRender, EventTypeDestroy:
		return false
	case EventTypeEnter, EventTypeMove, EventTypeExit, EventTypeClick,
		EventTypeRightClick, EventTypeDoubleClick, EventTypeDown, EventTypeUp,
		EventTypeRightDown, EventTypeRightUp, EventTypeMiss, EventTypeDragStart,
		EventTypeDrop, EventTypeDropEnter, EventTypeDropExit, EventTypeDragEnd,
		EventTypeScroll, EventTypeFocus, EventTypeBlur, EventTypeSubmit,
		EventTypeChange, EventTypeKeyDown, EventTypeKeyUp:
		return true
	default:
		return true
	}
}

func (ui *UI) ExecuteEvent(evtType EventType) bool {
	defer tracing.NewRegion("UI.ExecuteEvent").End()
	ui.events[evtType].Execute()
	return !ui.events[evtType].IsEmpty()
}

func (ui *UI) disabledEventBlocksSiblings(evtType EventType) bool {
	switch evtType {
	case EventTypeEnter, EventTypeMove, EventTypeClick, EventTypeRightClick,
		EventTypeDoubleClick, EventTypeDown, EventTypeUp, EventTypeRightDown,
		EventTypeRightUp, EventTypeDragStart, EventTypeDrop, EventTypeDropEnter,
		EventTypeDragEnd, EventTypeScroll:
		return true
	default:
		return false
	}
}

func (ui *UI) AddEvent(evtType EventType, evt func()) events.Id {
	return ui.events[evtType].Add(evt)
}

func (ui *UI) RemoveEvent(evtType EventType, evtId events.Id) {
	if evtId != 0 {
		ui.events[evtType].Remove(evtId)
	}
}

func (ui *UI) Event(evtType EventType) *events.Event {
	return &ui.events[evtType]
}

func (ui *UI) cleanDirty() { ui.dirtyType = DirtyTypeNone }

func labelDirtyRequiresRender(dirtyType DirtyType) bool {
	switch dirtyType {
	case DirtyTypeResize, DirtyTypeGenerated,
		DirtyTypeParentResize, DirtyTypeParentGenerated,
		DirtyTypeParentReGenerated:
		return true
	default:
		return false
	}
}

func (ui *UI) setDirtyInternal(dirtyType DirtyType) {
	defer tracing.NewRegion("UI.setDirtyInternal").End()
	if ui.IsType(ElementTypeLabel) {
		if labelDirtyRequiresRender(dirtyType) {
			ui.ToLabel().LabelData().renderRequired = true
		}
	} else {
		ui.ToPanel().PanelData().flags.setWasDirtied()
	}
	if ui.dirtyType == DirtyTypeNone || ui.dirtyType >= DirtyTypeParent || dirtyType == DirtyTypeGenerated {
		ui.dirtyType = dirtyType
		for i := 0; i < len(ui.entity.Children); i++ {
			kid := ui.entity.Children[i]
			all := AllOnEntity(kid)
			for _, cui := range all {
				if cui.dirty() == DirtyTypeNone || cui.dirty() > DirtyTypeParent {
					// TODO:  Let it know it was from the parent and what type
					if ui.dirtyType < DirtyTypeParent {
						cui.setDirtyInternal(DirtyTypeParent + ui.dirtyType)
					} else {
						cui.setDirtyInternal(ui.dirtyType)
					}
				}
			}
		}
	}
}

func (ui *UI) SetDirty(dirtyType DirtyType) {
	defer tracing.NewRegion("UI.SetDirty").End()
	if uiDiagEnabled && diagCapturing.Load() {
		diagRecord(ui, dirtyType)
	}
	ui.setDirtyInternal(dirtyType)
}

func (ui *UI) rootUI() *UI {
	defer tracing.NewRegion("UI.rootUI").End()
	root := &ui.entity
	var rootUI *UI = FirstOnEntity(root)
	for root.Parent != nil {
		if pui := FirstOnEntity(root.Parent); pui != nil {
			root = root.Parent
			rootUI = pui
		} else {
			break
		}
	}
	return rootUI
}

func (ui *UI) Clean() {
	defer tracing.NewRegion("UI.Clean").End()
	if ui.flags.dontClean() {
		return
	}
	var diagStart time.Time
	if uiDiagEnabled {
		diagStart = time.Now()
	}
	root := ui.rootUI()
	tree := []*UI{root}
	var createTree func(target *engine.Entity)
	createTree = func(target *engine.Entity) {
		for _, child := range target.Children {
			cui := FirstOnEntity(child)
			if cui != nil {
				tree = append(tree, cui)
				createTree(child)
			}
		}
	}
	createTree(root.Entity())
	stabilized := false
	maxIterations := 100
	iterations := 0
	// Convergence guard. Most layouts settle in 1-4 iterations. A few flex
	// subtrees never fully settle: they oscillate by a tiny amount forever
	// (the flex main-axis fit-content interaction plus the world<->local matrix
	// round-trip accumulating sub-pixel error across deep nesting). Without a
	// guard that burns all 100 iterations every frame. The rendered result is
	// already pixel-accurate, so after a minimum number of passes we stop once
	// the layout is only jittering below settleEpsilon, or once it is clearly
	// oscillating (the largest per-pass size change has stopped decreasing)
	// rather than converging.
	const settleEpsilon = 1.0
	const minGuardIterations = 10 // 10% of maxIterations
	prevMaxDelta := float32(-1)
	stalls := 0
	for !stabilized && maxIterations > 0 {
		stabilized = true
		maxDelta := float32(0)
		for i := range tree {
			if !tree[i].IsActive() {
				continue
			}
			before := tree[i].Layout().PixelSize()
			tree[i].cleanDirty()
			tree[i].Layout().update()
			tree[i].postLayoutUpdate()
			after := tree[i].Layout().PixelSize()
			if d := after.X() - before.X(); d > maxDelta {
				maxDelta = d
			} else if -d > maxDelta {
				maxDelta = -d
			}
			if d := after.Y() - before.Y(); d > maxDelta {
				maxDelta = d
			} else if -d > maxDelta {
				maxDelta = -d
			}
			stabilized = stabilized && tree[i].dirty() == DirtyTypeNone
		}
		maxIterations--
		iterations++
		if stabilized {
			break
		}
		if iterations >= minGuardIterations {
			if maxDelta < settleEpsilon {
				break // only sub-pixel jitter remains; accept it
			}
			if prevMaxDelta >= 0 && maxDelta >= prevMaxDelta*0.95 {
				// Not shrinking pass-over-pass -> oscillating, not converging.
				stalls++
				if stalls >= 2 {
					break
				}
			} else {
				stalls = 0
			}
		}
		prevMaxDelta = maxDelta
	}
	if uiDiagEnabled {
		if elapsed := time.Since(diagStart); !stabilized || elapsed > 30*time.Millisecond {
			uiCleanDiag(root, tree, iterations, elapsed, stabilized)
		}
	}
	for i := range tree {
		if !tree[i].IsActive() {
			continue
		}
		tree[i].GenerateScissor()
		tree[i].render()
	}
}

// uiCleanDiag logs why a Clean() pass was slow or failed to converge, naming the
// elements still dirty after the iteration cap so the oscillating subtree can be
// identified (dirty type, own size, parent size). TEMP diagnostic (uiDiagEnabled).
func uiCleanDiag(root *UI, tree []*UI, iterations int, elapsed time.Duration, stabilized bool) {
	n := uiDiagCount.Add(1)
	if n > 3 && n%60 != 0 {
		return // a few immediate samples, then throttle to limit log spam
	}
	// layoutMode: 0=Flow(block), 1=Grid, 2=Flex; -1=label/none.
	layoutModeOf := func(u *UI) int {
		if u == nil || u.IsType(ElementTypeLabel) {
			return -1
		}
		return u.ToPanel().PanelData().layoutMode
	}
	parentOf := func(t *UI) (string, int) {
		if p := t.entity.Parent; p != nil {
			pname := p.Name()
			if pu := FirstOnEntity(p); pu != nil {
				return pname, layoutModeOf(pu)
			}
			return pname, -1
		}
		return "<root>", -1
	}
	// 1) End-state stuck nodes (dirty != None after the 100-iteration cap) with
	// their own and their parent's layout mode (confirms flow vs flex).
	logged := 0
	for i := range tree {
		t := tree[i]
		if !t.IsActive() || t.dirty() == DirtyTypeNone || logged >= 12 {
			continue
		}
		logged++
		ps := t.Layout().PixelSize()
		pname, pmode := parentOf(t)
		slog.Warn("  ui-diag stuck",
			"name", t.entity.Name(), "elmType", int(t.elmType), "dirty", int(t.dirty()),
			"mode", layoutModeOf(t), "w", ps.X(), "h", ps.Y(),
			"parent", pname, "parentMode", pmode)
	}
	// 2) One extra instrumented settle-pass that captures WHICH call site
	// re-dirties each element (the actual non-convergence source). Since sizes
	// are stable, the culprit is a layout setter firing without a real change
	// (e.g. an exact-Approx guard tripping on a float artifact).
	diagDirtyCounts = make(map[string]int, 32)
	diagCapturing.Store(true)
	for i := range tree {
		t := tree[i]
		if !t.IsActive() {
			continue
		}
		t.cleanDirty()
		t.Layout().update()
		t.postLayoutUpdate()
	}
	diagCapturing.Store(false)
	type srcCount struct {
		what  string
		count int
	}
	srcs := make([]srcCount, 0, len(diagDirtyCounts))
	for k, c := range diagDirtyCounts {
		srcs = append(srcs, srcCount{k, c})
	}
	sort.Slice(srcs, func(i, j int) bool { return srcs[i].count > srcs[j].count })
	for i := range srcs {
		if i >= 12 {
			break
		}
		slog.Warn("  ui-diag dirty-source", "count", srcs[i].count, "what", srcs[i].what)
	}
	diagDirtyCounts = nil
	slog.Warn("UI.Clean slow/non-converged",
		"root", root.entity.Name(), "converged", stabilized, "iterations", iterations,
		"elapsedMs", elapsed.Milliseconds(), "treeSize", len(tree), "dirtyEvents", len(srcs))
}

func (ui *UI) GenerateScissor() {
	target := &ui.entity.Transform
	pos := target.WorldPosition()
	size := target.WorldScale()
	bounds := matrix.Vec4{
		pos.X() - size.X()*0.5,
		pos.Y() - size.Y()*0.5,
		pos.X() + size.X()*0.5,
		pos.Y() + size.Y()*0.5,
	}
	if !ui.IsType(ElementTypeLabel) {
		outset := ui.ToPanel().OutlineOutset()
		bounds.SetX(bounds.X() - outset)
		bounds.SetY(bounds.Y() - outset)
		bounds.SetZ(bounds.Z() + outset)
		bounds.SetW(bounds.W() + outset)
	}
	if !ui.entity.IsRoot() {
		p := FirstPanelOnEntity(ui.entity.Parent)
		for p.PanelData().overflow == OverflowVisible && !p.entity.IsRoot() {
			p = FirstPanelOnEntity(p.entity.Parent)
		}
		//if !p.entity.IsRoot() {
		ps := p.Base().selfScissor()
		bounds.SetX(max(bounds.X(), ps.X()))
		bounds.SetY(max(bounds.Y(), ps.Y()))
		bounds.SetZ(min(bounds.Z(), ps.Z()))
		bounds.SetW(min(bounds.W(), ps.W()))
		//}
	}
	ui.setScissor(bounds)
}

func (ui *UI) setScissor(scissor matrix.Vec4) {
	ui.setScissorInternal(scissor)
}

func (ui *UI) setScissorInternal(scissor matrix.Vec4) {
	if ui.shaderData.Scissor.Equals(scissor) {
		return
	}
	for i := 0; i < len(ui.entity.Children); i++ {
		cUI := FirstOnEntity(ui.entity.Children[i])
		if cUI != nil {
			cUI.setScissorInternal(scissor)
		}
	}
	ui.shaderData.Scissor = scissor
	me := FirstOnEntity(&ui.entity)
	if me.elmType == ElementTypeLabel {
		ld := me.ToLabel().LabelData()
		for i := range ld.runeDrawings {
			ld.runeDrawings[i].ShaderData.(*rendering.TextShaderData).Scissor = scissor
		}
	}
}

func (ui *UI) requestEvent(evtType EventType) bool {
	defer tracing.NewRegion("UI.requestEvent").End()
	if ui.disabledBlocksEvent(evtType) {
		return ui.disabledEventBlocksSiblings(evtType)
	}
	if ui.events[evtType].IsEmpty() {
		return false
	}
	man := ui.man.Value()
	if man != nil {
		man.Group.requestEvent(ui, evtType)
	} else {
		ui.ExecuteEvent(evtType)
	}
	return true
}

func (ui *UI) eventUpdates() {
	defer tracing.NewRegion("UI.eventUpdates").End()
	host := ui.man.Value().Host
	cursor := &host.Window.Cursor
	mouse := &host.Window.Mouse
	if cursor.Moved() {
		pos := ui.cursorPos(cursor)
		ui.containedCheck(cursor, &ui.entity)
		if ui.IsDisabled() {
			return
		}
		if ui.flags.isDown() && !ui.flags.drag() {
			w := ui.Host().Window.Width()
			h := ui.Host().Window.Height()
			wmm, hmm, _ := host.Window.SizeMM()
			threshold := max(windowing.DPI2PX(w, wmm, 1), windowing.DPI2PX(h, hmm, 1))
			if ui.downPos.Distance(pos) > float32(threshold) {
				ui.dragStartPos = ui.entity.Transform.WorldPosition()
				ui.flags.setDrag()
				ui.requestEvent(EventTypeDragStart)
			}
		}
	}
	if cursor.Pressed() {
		ui.containedCheck(cursor, &ui.entity)
		if ui.IsDisabled() {
			return
		}
		if ui.flags.hovering() && !ui.flags.isDown() {
			ui.flags.setIsDown()
			ui.downPos = ui.cursorPos(cursor)
			ui.requestEvent(EventTypeDown)
			ui.flags.setCantMiss()
		} else if !ui.flags.hovering() && !ui.flags.cantMiss() {
			ui.requestEvent(EventTypeMiss)
		} else {
			ui.flags.resetCantMiss()
		}
	}
	if mouse.Pressed(hid.MouseButtonRight) {
		ui.containedCheck(cursor, &ui.entity)
		if ui.IsDisabled() {
			return
		}
		if ui.flags.hovering() && !ui.flags.isRightDown() {
			ui.flags.setIsRightDown()
			ui.downPos = ui.cursorPos(cursor)
			ui.requestEvent(EventTypeRightDown)
			ui.flags.setCantMiss()
		}
	}
	if cursor.Released() {
		if ui.IsDisabled() {
			ui.flags.resetIsDown()
			ui.flags.resetDrag()
			ui.flags.resetCantMiss()
			return
		}
		if ui.flags.hovering() {
			ui.requestEvent(EventTypeUp)
			if windowing.HasDragData() {
				ui.requestEvent(EventTypeDrop)
			}
		}
		if ui.flags.lastActive() {
			if ui.flags.isDown() {
				ui.flags.resetIsDown()
				dragged := false
				if ui.flags.drag() {
					p := ui.entity.Transform.WorldPosition()
					dragged = ui.dragStartPos.Distance(p) > 5
				}
				ui.flags.resetDrag()
				ui.requestEvent(EventTypeDragEnd)
				if ui.flags.hovering() && !dragged {
					rt := host.Runtime()
					if rt-ui.lastClick < dblCLickTime && !ui.events[EventTypeDoubleClick].IsEmpty() {
						ui.requestEvent(EventTypeDoubleClick)
						ui.lastClick = 0
					} else {
						ui.requestEvent(EventTypeClick)
						ui.lastClick = rt
					}
				}
			}
			ui.flags.resetCantMiss()
		}
	}
	if mouse.Released(hid.MouseButtonRight) {
		if ui.IsDisabled() {
			ui.flags.resetIsRightDown()
			return
		}
		if ui.flags.hovering() {
			ui.requestEvent(EventTypeUp)
			if ui.flags.isRightDown() {
				ui.requestEvent(EventTypeRightClick)
			}
		}
		ui.flags.resetIsRightDown()
	}
	if mouse.Scrolled() && ui.flags.hovering() {
		if ui.IsDisabled() {
			return
		}
		ui.requestEvent(EventTypeScroll)
	}
}

func (ui *UI) Update(deltaTime float64) {
	defer tracing.NewRegion("UI.Update").End()
	// TODO:  Everything should be clean by this point, there is a bug where
	// by the time the wait group in ui_manager.go:~49 is done, something is
	// still in-flight to be cleaned?
	//if ui.dirtyType != DirtyTypeNone {
	//	ui.Clean()
	//}
	if ui.entity.IsActive() {
		ui.flags.setLastActive()
	} else {
		ui.flags.resetLastActive()
	}
}

func (ui *UI) cursorPos(cursor *hid.Cursor) matrix.Vec2 {
	defer tracing.NewRegion("UI.cursorPos").End()
	pos := cursor.Position()
	host := ui.man.Value().Host
	pos[matrix.Vx] -= matrix.Float(host.Window.Width()) * 0.5
	pos[matrix.Vy] -= matrix.Float(host.Window.Height()) * 0.5
	return pos
}

func (ui *UI) containedCheck(cursor *hid.Cursor, entity *engine.Entity) {
	defer tracing.NewRegion("UI.containedCheck").End()
	cp := ui.cursorPos(cursor)
	contained := entity.Transform.ContainsPoint2D(cp)
	if contained && ui.hasScissor() {
		contained = ui.shaderData.Scissor.ScreenAreaContains(cp.X(), cp.Y())
	}
	if !ui.flags.hovering() && contained {
		ui.flags.setHovering()
		if ui.IsDisabled() {
			return
		}
		// This is to resolve the parent not getting it's exit call when the
		// cursor enters a child element, effectively taking focus from the
		// parent
		if ui.requestEvent(EventTypeEnter) && ui.entity.Parent != nil {
			FirstOnEntity(ui.entity.Parent).requestEvent(EventTypeExit)
		}
		if windowing.HasDragData() {
			if ui.requestEvent(EventTypeDropEnter) && ui.entity.Parent != nil {
				FirstOnEntity(ui.entity.Parent).requestEvent(EventTypeDropExit)
			}
		}
	} else if ui.flags.hovering() && !contained {
		ui.flags.resetHovering()
		if ui.IsDisabled() {
			return
		}
		ui.requestEvent(EventTypeExit)
		// This is to resolve the parent not getting enter call when the
		// cursor exits a child element puttin focus back on the parent
		if !ui.events[EventTypeEnter].IsEmpty() && ui.entity.Parent != nil {
			FirstOnEntity(ui.entity.Parent).flags.resetHovering()
		}
		if windowing.HasDragData() {
			ui.requestEvent(EventTypeDropExit)
		}
	} else if ui.flags.hovering() && contained {
		if ui.IsDisabled() {
			return
		}
		ui.requestEvent(EventTypeMove)
	}
}

func (ui *UI) changed() {
	defer tracing.NewRegion("UI.changed").End()
	ui.ExecuteEvent(EventTypeChange)
}

func (ui *UI) layoutChanged(dirtyType DirtyType) {
	defer tracing.NewRegion("UI.layoutChanged").End()
	ui.SetDirty(dirtyType)
}

func (ui *UI) cleanIfNeeded() {
	defer tracing.NewRegion("UI.cleanIfNeeded").End()
	if ui.anyChildDirty() {
		ui.Clean()
	}
}

func (ui *UI) anyChildDirty() bool {
	defer tracing.NewRegion("UI.anyChildDirty").End()
	if !ui.IsActive() || ui.entity.IsDestroyed() {
		return false
	}
	if ui.dirtyType != DirtyTypeNone {
		return true
	}
	for i := range ui.entity.Children {
		cui := FirstOnEntity(ui.entity.Children[i])
		if cui != nil && cui.anyChildDirty() {
			return true
		}
	}
	return false
}

func (ui *UI) updateFromManager(deltaTime float64) {
	defer tracing.NewRegion("UI.updateFromManager").End()
	if !ui.IsActive() {
		return
	}
	switch ui.elmType {
	case ElementTypeInput:
		ui.ToInput().update(deltaTime)
	case ElementTypeTextArea:
		ui.ToTextArea().update(deltaTime)
	case ElementTypeLabel:
		ui.Update(deltaTime)
	case ElementTypePanel:
		ui.ToPanel().update(deltaTime)
	case ElementTypeButton:
		ui.ToPanel().update(deltaTime)
	case ElementTypeSelect:
		ui.ToSelect().update(deltaTime)
	case ElementTypeSlider:
		ui.ToSlider().update(deltaTime)
	case ElementTypeImage:
		ui.ToImage().update(deltaTime)
	case ElementTypeCheckbox:
		ui.ToPanel().update(deltaTime)
	}
}

func (ui *UI) Show() {
	defer tracing.NewRegion("UI.Show").End()
	ui.entity.Activate()
}

func (ui *UI) Hide() {
	defer tracing.NewRegion("UI.Hide").End()
	ui.entity.Deactivate()
}

func (ui *UI) SetVisibility(visible bool) {
	defer tracing.NewRegion("UI.ShowToggle").End()
	if visible {
		ui.entity.Activate()
	} else {
		ui.entity.Deactivate()
	}
}

func (ui *UI) FindByName(name string) *UI {
	defer tracing.NewRegion("UI.FindByName").End()
	e := ui.entity.FindByName(name)
	if e != nil {
		return FirstOnEntity(e)
	}
	return nil
}

func (ui *UI) IsInFrontOf(other *UI) bool {
	defer tracing.NewRegion("UI.IsInFrontOf").End()
	if ui == other {
		return true
	}
	return ui.entity.Transform.WorldPosition().Z() >
		other.entity.Transform.WorldPosition().Z()
}

func (ui *UI) Clone(parent *engine.Entity) *UI {
	cpy := ui.man.Value().Add()
	switch ui.elmType {
	case ElementTypeLabel:
		ui.ToLabel().Clone(cpy.ToLabel())
	case ElementTypePanel:
		t := ui.ToPanel()
		cpy.ToPanel().Init(t.Background(), ElementTypePanel)
	case ElementTypeButton:
		t := ui.ToButton()
		cpy.ToButton().Init(ui.ToPanel().Background(), t.Label().Text())
	case ElementTypeCheckbox:
		cpy.ToCheckbox().Init()
	case ElementTypeImage:
		t := ui.ToImage()
		tData := t.ImageData()
		if len(tData.flipBook) > 0 {
			cpy.ToImage().InitFlipbook(tData.fps, tData.flipBook)
		} else if tData.spriteSheet.IsValid() {
			s, _ := tData.spriteSheet.ToJson()
			cpy.ToImage().InitSpriteSheet(tData.fps, ui.ToPanel().Background(), s)
		} else {
			bg := ui.ToPanel().Background()
			if bg != nil {
				cpy.ToImage().Init(bg)
			} else {
				slog.Error("failed to clone image UI: no texture source was available")
				cpy.ToImage().Init(nil)
			}
		}
	case ElementTypeInput:
		t := ui.ToInput()
		tData := t.InputData()
		cpyInput := cpy.ToInput()
		cpyInput.Init(tData.placeholder.Text())
		cpyInput.SetType(tData.inputType)
		cpyInput.SetRequired(tData.required)
		cpyInput.SetTextWithoutEvent(t.Text())
	case ElementTypeTextArea:
		t := ui.ToTextArea()
		tData := t.Data()
		cpyTextArea := cpy.ToTextArea()
		cpyTextArea.Init(tData.placeholder.Text())
		cpyTextArea.SetRequired(tData.required)
		cpyTextArea.SetTextWithoutEvent(t.Text())
	case ElementTypeProgressBar:
		t := ui.ToProgressBar()
		cpy.ToProgressBar().Init(t.data().fgPanel.Background(), ui.ToPanel().Background())
	case ElementTypeSelect:
		t := ui.ToSelect()
		cpy.ToSelect().Init(t.SelectData().text, t.SelectData().options)
	case ElementTypeSlider:
		cpy.ToSlider().Init()
	}
	cpy.SetDisabled(ui.IsDisabled())
	if parent != nil {
		panel := FirstPanelOnEntity(parent)
		if panel != nil {
			panel.AddChild(cpy)
		} else {
			cpy.entity.SetParent(parent)
		}
	}
	cpy.entity.Transform.Copy(ui.entity.Transform)
	cpy.SetDirty(DirtyTypeGenerated)
	return cpy
}
