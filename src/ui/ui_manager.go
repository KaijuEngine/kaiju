package ui

import (
	"kaiju/engine"
	"kaiju/pooling"
	"sync"
)

const (
	concurrentUpdateLimit = 100
)

type Manager struct {
	Host     *engine.Host
	Group    *Group
	pools    pooling.PoolGroup[UI]
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
		if (*elm).Entity().IsRoot() {
			roots = append(roots, elm)
		} else {
			children = append(children, elm)
		}
	})
	// First we update all the root UI elements, this will stabilize the tree

	waitForLimit := make(chan struct{}, concurrentUpdateLimit)
	for range concurrentUpdateLimit {
		waitForLimit <- struct{}{}
	}

	wg.Add(len(roots))
	threads := man.Host.Threads()
	for i := range roots {
		threads.AddWork(func() {
			roots[i].cleanIfNeeded()
			wg.Done()
		})
	}
	wg.Wait()
	// Then we go through and update all the remaining UI elements
	all := append(children, roots...)
	wg.Add(len(all))
	for i := range all {
		threads.AddWork(func() {
			all[i].updateFromManager(deltaTime)
			wg.Done()
		})
	}
	wg.Wait()
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
	return ui
}

func (man *Manager) Remove(ui *UI) {
	man.pools.Remove(ui.poolId, ui.id)
}

func (man *Manager) Reserve(additionalElements int) {
	man.pools.Reserve(additionalElements)
	man.Host.ReserveEntities(additionalElements)
}
