/******************************************************************************/
/* ui_manager.go                                                              */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ui

import (
	"kaiju/engine"
	"kaiju/engine/pooling"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"sync"
)

const (
	concurrentUpdateLimit = 100
)

type Manager struct {
	Host     *engine.Host
	Group    *Group
	pools    pooling.PoolGroup[UI]
	hovered  [][]*UI
	updateId int
}

type manUp struct {
	deltaTime float64
	ui        *UI
}

func (man *Manager) update(deltaTime float64) {
	defer tracing.NewRegion("ui.Manager.update").End()
	// There is no update without a host, this is safe
	wg := sync.WaitGroup{}
	roots := []*UI{}
	children := []*UI{}
	man.pools.Each(func(elm *UI) {
		if elm.entity.IsDestroyed() {
			return
		}
		if elm.entity.IsRoot() {
			roots = append(roots, elm)
		} else {
			children = append(children, elm)
		}
	})
	// First we update all the root UI elements, this will stabilize the tree
	wg.Add(len(roots))
	threads := man.Host.Threads()
	for i := range roots {
		threads.AddWork(func(int) {
			roots[i].cleanIfNeeded()
			wg.Done()
		})
	}
	wg.Wait()
	// Then we go through and update all the remaining UI elements
	all := append(children, roots...)
	wg.Add(len(all))
	tCount := threads.ThreadCount()
	if len(man.hovered) != tCount {
		man.hovered = make([][]*UI, tCount)
	} else {
		for i := range len(man.hovered) {
			man.hovered[i] = klib.WipeSlice(man.hovered[i])
		}
	}
	for i := range all {
		threads.AddWork(func(threadId int) {
			e := all[i]
			e.updateFromManager(deltaTime)
			if e.isActive() && e.hovering && e.elmType == ElementTypePanel && e.ToPanel().Background() != nil {
				man.hovered[threadId] = append(man.hovered[threadId], e)
			}
			wg.Done()
		})
	}
	wg.Wait()
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
	man.updateId = host.UIUpdater.AddUpdate(man.update)
	man.Group = NewGroup()
	man.Group.Attach(man.Host)
	man.Group.SetThreaded()
}

func (man *Manager) Release() {
	defer tracing.NewRegion("ui.Manager.Release").End()
	man.Clear()
	man.Host.UIUpdater.RemoveUpdate(man.updateId)
	man.Group.Detach(man.Host)
	man.updateId = 0
}

func (man *Manager) Clear() {
	defer tracing.NewRegion("ui.Manager.Clear").End()
	man.pools.Each(func(ui *UI) { ui.Entity().Destroy() })
	// Clearing the pools shouldn't be needed as destroying the entities
	// will remove the entry from the pool
}

func (man *Manager) Add() *UI {
	defer tracing.NewRegion("ui.Manager.Add").End()
	ui, poolId, elmId := man.pools.Add()
	*ui = UI{
		poolId: poolId,
		id:     elmId,
		man:    man,
	}
	ui.entity.Init(man.Host.WorkGroup())
	man.Host.AddEntity(&ui.entity)
	return ui
}

func (man *Manager) Remove(ui *UI) {
	defer tracing.NewRegion("ui.Manager.Remove").End()
	id := ui.id
	pid := ui.poolId
	man.pools.Remove(pid, id)
}

func (man *Manager) Reserve(additionalElements int) {
	defer tracing.NewRegion("ui.Manager.Reserve").End()
	man.pools.Reserve(additionalElements)
	man.Host.ReserveEntities(additionalElements)
}
