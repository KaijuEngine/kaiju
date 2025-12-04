/******************************************************************************/
/* updater.go                                                                 */
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

package engine

import (
	"kaiju/klib"
	"kaiju/platform/concurrent"
	"kaiju/platform/profiler/tracing"
	"sync"
	"sync/atomic"
)

type UpdateId int

type engineUpdate struct {
	id     UpdateId
	update func(float64)
}

// Updater is a struct that stores update functions to be called when the
// #Updater.Update function is called. This simply goes through the list
// from top to bottom and calls each function.
//
// *Note that update functions are unordered, so don't rely on the order*
type Updater struct {
	updates    map[UpdateId]engineUpdate
	threads    *concurrent.Threads
	backAdd    []engineUpdate
	backRemove []UpdateId
	nextId     atomic.Int32
	lastDelta  float64
}

// IsConcurrent will return if this updater is a concurrent updater
func (u *Updater) IsConcurrent() bool {
	return u.threads != nil
}

// NewUpdater creates a new #Updater struct and returns it
func NewUpdater() Updater {
	return Updater{updates: make(map[UpdateId]engineUpdate)}
}

// NewConcurrentUpdater creates a new concurrent #Updater struct and returns it
func NewConcurrentUpdater(threads *concurrent.Threads) Updater {
	u := NewUpdater()
	u.threads = threads
	return u
}

// AddUpdate adds an update function to the list of updates to be called when
// the #Updater.Update function is called. It returns the id of the update
// function that was added so that it can be removed later.
//
// The update function is added to a back-buffer so it will not begin updating
// until the next call to #Updater.Update.
func (u *Updater) AddUpdate(update func(float64)) UpdateId {
	id := UpdateId(u.nextId.Add(1))
	u.backAdd = append(u.backAdd, engineUpdate{
		id:     id,
		update: update,
	})
	return id
}

// RemoveUpdate removes an update function from the list of updates to be called
// when the #Updater.Update function is called. It takes the id of the update
// function that was returned when the update function was added.
//
// The update function is removed from a back-buffer so it will not be removed
// until the next call to #Updater.Update.
func (u *Updater) RemoveUpdate(id *UpdateId) {
	if *id > 0 {
		u.backRemove = append(u.backRemove, *id)
	}
	id.reset()
}

// Update calls all of the update functions that have been added to the updater.
// It takes a deltaTime parameter that is the approximate amount of time since
// the last call to #Updater.Update.
func (u *Updater) Update(deltaTime float64) {
	defer tracing.NewRegion("Updater.Update").End()
	u.lastDelta = deltaTime
	u.addInternal()
	u.removeInternal()
	if u.IsConcurrent() {
		u.concurrentUpdate(deltaTime)
	} else {
		u.inlineUpdate(deltaTime)
	}
}

// Destroy cleans up the updater and should be called when the updater is no
// longer needed. It will close the pending and complete channels and clear the
// updates map.
func (u *Updater) Destroy() {
	clear(u.updates)
	u.backAdd = make([]engineUpdate, 0)
	u.backRemove = make([]UpdateId, 0)
}

func (u *Updater) inlineUpdate(deltaTime float64) {
	for i := range u.updates {
		u.updates[i].update(deltaTime)
	}
}

func (u *Updater) concurrentUpdate(deltaTime float64) {
	work := make([]func(int), 0, len(u.updates))
	wg := sync.WaitGroup{}
	for _, v := range u.updates {
		wg.Add(1)
		work = append(work, func(int) {
			v.update(deltaTime)
			wg.Done()
		})
	}
	u.threads.AddWork(work)
	wg.Wait()
}

func (u *Updater) addInternal() {
	for _, update := range u.backAdd {
		u.updates[update.id] = update
	}
	u.backAdd = klib.WipeSlice(u.backAdd)
}

func (u *Updater) removeInternal() {
	for _, id := range u.backRemove {
		delete(u.updates, id)
	}
	u.backRemove = klib.WipeSlice(u.backRemove)
}

func (u *UpdateId) reset() { *u = 0 }
