package ui

import (
	"kaiju/engine"
	"kaiju/engine/pooling"
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
			man.hovered[i] = man.hovered[i][:0]
		}
	}
	for i := range all {
		threads.AddWork(func(threadId int) {
			e := all[i]
			e.updateFromManager(deltaTime)
			if e.hovering && e.elmType == ElementTypePanel && e.ToPanel().Background() != nil {
				man.hovered[threadId] = append(man.hovered[threadId], e)
			}
			wg.Done()
		})
	}
	wg.Wait()
}

func (man *Manager) Hovered() []*UI {
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
	man.Host = host
	man.updateId = host.UIUpdater.AddUpdate(man.update)
	man.Group = NewGroup()
	man.Group.Attach(host)
	man.Group.SetThreaded()
}

func (man *Manager) Release() {
	man.Clear()
	man.Host.UIUpdater.RemoveUpdate(man.updateId)
	man.updateId = 0
	man.Group.Detach(man.Host)
}

func (man *Manager) Clear() {
	man.pools.Each(func(ui *UI) { ui.Entity().Destroy() })
	// Clearing the pools shouldn't be needed as destroying the entities
	// will remove the entry from the pool
}

func (man *Manager) Add() *UI {
	ui, poolId, elmId := man.pools.Add()
	*ui = UI{}
	ui.poolId = poolId
	ui.id = elmId
	ui.man = man
	ui.group = man.Group
	ui.entity.Init(man.Host.WorkGroup())
	man.Host.AddEntity(&ui.entity)
	return ui
}

func (man *Manager) Remove(ui *UI) {
	man.pools.Remove(ui.poolId, ui.id)
}

func (man *Manager) Reserve(additionalElements int) {
	man.pools.Reserve(additionalElements)
	man.Host.ReserveEntities(additionalElements)
}
