/******************************************************************************/
/* ui_manager.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import (
	"runtime"
	"sync"
	"weak"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
)

type Manager struct {
	Host              *engine.Host
	Group             Group
	pools             pooling.PoolGroup[UI]
	hovered           [][]*UI
	itrRoots          []*UI
	itrChildren       []*UI
	itrAll            []*UI
	updateId          engine.UpdateId
	skipUpdate        int
	dirtyBatchDepth   int
	dirtyBatchTargets []*UI
	resizeEvtId       events.Id
	windowResized     bool
	windowMinimized   bool
}

func (man *Manager) update(deltaTime float64) {
	defer tracing.NewRegion("ui.Manager.update").End()
	if man.windowMinimized || (man.skipUpdate > 0 && !man.windowResized) {
		return
	}
	man.itrRoots = klib.WipeSlice(man.itrRoots)
	man.itrChildren = klib.WipeSlice(man.itrChildren)
	man.itrAll = klib.WipeSlice(man.itrAll)
	// There is no update without a host, this is safe
	wg := sync.WaitGroup{}
	man.pools.Each(func(elm *UI) {
		if elm.entity.IsDestroyed() {
			return
		}
		if elm.entity.IsRoot() {
			man.itrRoots = append(man.itrRoots, elm)
		} else {
			man.itrChildren = append(man.itrChildren, elm)
		}
	})
	// First we update all the root UI elements, this will stabilize the tree
	wg.Add(len(man.itrRoots))
	threads := man.Host.UIThreads()
	work := make([]func(int), len(man.itrRoots))
	for i := range man.itrRoots {
		work[i] = func(int) {
			man.itrRoots[i].cleanIfNeeded()
			wg.Done()
		}
	}
	threads.AddWork(work)
	wg.Wait()
	// Then we go through and update all the remaining UI elements
	man.itrAll = append(man.itrAll, man.itrChildren...)
	man.itrAll = append(man.itrAll, man.itrRoots...)
	wg.Add(len(man.itrAll))
	tCount := threads.ThreadCount()
	if len(man.hovered) != tCount {
		man.hovered = make([][]*UI, tCount)
	} else {
		for i := range len(man.hovered) {
			man.hovered[i] = klib.WipeSlice(man.hovered[i])
		}
	}
	work = make([]func(int), len(man.itrAll))
	for i := range man.itrAll {
		work[i] = func(threadId int) {
			e := man.itrAll[i]
			e.updateFromManager(deltaTime)
			if e.IsActive() && e.flags.hovering() && !e.IsType(ElementTypeLabel) && e.ToPanel().Background() != nil {
				man.hovered[threadId] = append(man.hovered[threadId], e)
			}
			wg.Done()
		}
	}
	threads.AddWork(work)
	wg.Wait()
	man.windowResized = false
}

func (man *Manager) Hovered() []*UI {
	defer tracing.NewRegion("ui.Manager.Hovered").End()
	count := 0
	for i := range man.hovered {
		count += len(man.hovered[i])
	}
	out := make([]*UI, 0, count)
	for i := range man.hovered {
		for j := range man.hovered[i] {
			out = append(out, man.hovered[i][j])
		}
	}
	return out
}

func (man *Manager) Init(host *engine.Host) {
	defer tracing.NewRegion("ui.Manager.Init").End()
	man.Host = host
	wMan := weak.Make(man)
	man.updateId = host.UIUpdater.AddUpdate(func(deltaTime float64) {
		if wMan.Value() != nil {
			wMan.Value().update(deltaTime)
		}
	})
	man.Group.Attach(man.Host)
	man.Group.SetThreaded()
	man.resizeEvtId = host.Window.OnResize.Add(func() {
		if m := wMan.Value(); m != nil {
			m.windowResized = true
			m.windowMinimized = m.Host.Window.IsMinimized()
		}
	})
	type manCleanup struct {
		host          weak.Pointer[engine.Host]
		win           weak.Pointer[windowing.Window]
		updateId      engine.UpdateId
		resizeId      events.Id
		groupUpdateId engine.UpdateId
	}
	clean := manCleanup{weak.Make(host), weak.Make(host.Window),
		man.updateId, man.resizeEvtId, man.Group.updateId}
	runtime.AddCleanup(man, func(c manCleanup) {
		h := c.host.Value()
		if h == nil {
			return
		}
		h.UIUpdater.RemoveUpdate(&c.updateId)
		h.UILateUpdater.RemoveUpdate(&c.groupUpdateId)
		w := c.win.Value()
		if w == nil {
			return
		}
		w.OnResize.Remove(c.resizeId)
	}, clean)
}

func (man *Manager) Clear() {
	defer tracing.NewRegion("ui.Manager.Clear").End()
	man.pools.Each(func(ui *UI) { man.Host.DestroyEntity(ui.Entity()) })
	// Clearing the pools shouldn't be needed as destroying the entities
	// will remove the entry from the pool
}

// Shutdown synchronously detaches the Manager from its host: removes its
// update from UIUpdater, detaches its Group from UILateUpdater, and removes
// its window-resize handler. It also clears any UI elements still owned by
// the Manager so the next Init starts clean.
//
// This is safe to call when Manager hasn't been Init'd yet (no-op). It is
// also idempotent — calling it twice in a row tears down nothing the second
// time. After Shutdown the Manager can be Init'd again on the same or a
// different host.
//
// Without this, calling Init twice on the same Manager (which happens when
// a workspace is disabled then re-enabled in the editor) leaves the
// previous update callback registered, so two goroutines race on the same
// Manager's iteration slices and panic with index-out-of-range.
func (man *Manager) Shutdown() {
	defer tracing.NewRegion("ui.Manager.Shutdown").End()
	if man.Host == nil {
		return
	}
	man.Clear()
	man.Group.Detach(man.Host)
	man.Host.UIUpdater.RemoveUpdate(&man.updateId)
	if man.Host.Window != nil {
		man.Host.Window.OnResize.Remove(man.resizeEvtId)
	}
	man.Host = nil
}

func (man *Manager) Add() *UI {
	defer tracing.NewRegion("ui.Manager.Add").End()
	ui, poolId, elmId := man.pools.Add()
	*ui = UI{
		poolId: poolId,
		id:     elmId,
		man:    weak.Make(man),
	}
	ui.entity.Init(man.Host.WorkGroup())
	return ui
}

// RunDirtyBatch batches UI dirty propagation for construction paths that live
// outside this package. It is intended for Kaiju's markup/control builders, not
// as a general gameplay-facing API.
func RunDirtyBatch(man *Manager, fn func()) {
	if man == nil {
		fn()
		return
	}
	man.runDirtyBatch(fn)
}

func (man *Manager) runDirtyBatch(fn func()) {
	man.beginDirtyBatch()
	defer man.endDirtyBatch()
	fn()
}

func (man *Manager) beginDirtyBatch() {
	man.dirtyBatchDepth++
}

func (man *Manager) endDirtyBatch() {
	if man.dirtyBatchDepth == 0 {
		return
	}
	man.dirtyBatchDepth--
	if man.dirtyBatchDepth == 0 {
		man.flushDirtyBatch()
	}
}

func (man *Manager) isDirtyBatching() bool {
	return man != nil && man.dirtyBatchDepth > 0
}

func (man *Manager) recordDirtyBatchTarget(target *UI) {
	if !man.isDirtyBatching() || target == nil {
		return
	}
	man.dirtyBatchTargets = append(man.dirtyBatchTargets, target)
}

func (man *Manager) flushDirtyBatch() {
	if len(man.dirtyBatchTargets) == 0 {
		return
	}
	roots := make([]*UI, 0, len(man.dirtyBatchTargets))
	for i := range man.dirtyBatchTargets {
		root := man.dirtyBatchTargets[i].dirtyBatchRoot()
		if root != nil && root.IsActive() && !root.entity.IsDestroyed() {
			roots = appendDirtyBatchRoot(roots, root)
		}
	}
	man.dirtyBatchTargets = man.dirtyBatchTargets[:0]
	recordDirtyBatchFlush(len(roots))
	flags := uiDirtyStyle | uiDirtyLayoutSelf | uiDirtyLayoutChildren | uiDirtyScissor | uiDirtyRender
	for i := range roots {
		roots[i].addDirtyFlags(flags)
		roots[i].bubbleDirty(flags)
	}
}

func appendDirtyBatchRoot(roots []*UI, root *UI) []*UI {
	for i := range roots {
		if roots[i] == root || root.entity.HasParent(&roots[i].entity) {
			return roots
		}
		if roots[i].entity.HasParent(&root.entity) {
			roots[i] = root
			return roots
		}
	}
	return append(roots, root)
}

func (man *Manager) Remove(ui *UI) {
	defer tracing.NewRegion("ui.Manager.Remove").End()
	id := ui.id
	pid := ui.poolId
	man.pools.Remove(pid, id)
	ui.layout.Stylizer = nil
}

func (man *Manager) Reserve(additionalElements int) {
	defer tracing.NewRegion("ui.Manager.Reserve").End()
	man.pools.Reserve(additionalElements)
}

func (man *Manager) IsUpdateDisabled() bool { return man.skipUpdate > 0 }

func (man *Manager) DisableUpdate() {
	man.Host.RunNextFrame(func() { man.skipUpdate++ })
}

func (man *Manager) EnableUpdate() { man.skipUpdate = max(0, man.skipUpdate-1) }
