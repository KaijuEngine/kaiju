package ui

import (
	"kaiju/engine"
	"kaiju/pooling"
	"sync"
)

type Manager struct {
	host     *engine.Host
	group    *Group
	pools    pooling.PoolGroup[UI]
	updateId int
}

type manUp struct {
	deltaTime float64
	ui        *UI
}

func (man *Manager) update(deltaTime float64) {
	wg := sync.WaitGroup{}
	// First we update all the root UI elements, this will stabilize the tree
	man.pools.Each(func(elm *UI) {
		if (*elm).Entity().IsRoot() {
			wg.Add(1)
			go func() {
				elm.rootCleanIfNeeded()
				wg.Done()
			}()
		}
	})
	wg.Wait()
	// Then we go through and update all the remaining UI elements
	wg = sync.WaitGroup{}
	man.pools.Each(func(elm *UI) {
		if !(*elm).Entity().IsRoot() {
			wg.Add(1)
			go func() {
				elm.updateFromManager(deltaTime)
				wg.Done()
			}()
		}
	})
	wg.Wait()
}

func (man *Manager) Init(host *engine.Host) {
	man.host = host
	man.updateId = host.UIUpdater.AddUpdate(man.update)
	man.group = NewGroup()
	man.group.Attach(host)
	man.group.SetThreaded()
}

func (man *Manager) Release() {
	man.Clear()
	man.host.UIUpdater.RemoveUpdate(man.updateId)
	man.updateId = 0
	man.group.Detach(man.host)
}

func (man *Manager) Clear() {
	man.pools.Each(func(ui *UI) { ui.Entity().Destroy() })
	// Clearing the pools shouldn't be needed as destroying the entities
	// will remove the entry from the pool
}

func (man *Manager) Add() *UI {
	ui, poolId, elmId := man.pools.Add()
	ui.poolId = poolId
	ui.id = elmId
	ui.man = man
	ui.SetGroup(man.group)
	return ui
}

func (man *Manager) Remove(ui *UI) {
	man.pools.Remove(ui.poolId, ui.id)
}

func (man *Manager) Reserve(additionalElements int) {
	man.pools.Reserve(additionalElements)
	man.host.ReserveEntities(additionalElements)
}
