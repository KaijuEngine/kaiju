/******************************************************************************/
/* updater.go                                                                 */
/******************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/******************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/******************************************************************************/

package engine

type engineUpdate struct {
	id     int
	update func(float64)
}

// Updater is a struct that stores update functions to be called when the
// #Updater.Update function is called. This simply goes through the list
// from top to bottom and calls each function.
//
// *Note that update functions are unordered, so don't rely on the order*
type Updater struct {
	updates      map[int]engineUpdate
	backAdd      []engineUpdate
	backRemove   []int
	nextId       int
	lastDelta    float64
	pending      chan int
	complete     chan int
	isConcurrent bool
}

// NewUpdater creates a new #Updater struct and returns it
func NewUpdater() Updater {
	return Updater{
		updates:      make(map[int]engineUpdate),
		backAdd:      make([]engineUpdate, 0),
		backRemove:   make([]int, 0),
		nextId:       1,
		pending:      make(chan int, 100),
		complete:     make(chan int, 100),
		isConcurrent: false,
	}
}

// StartConcurrent starts the number of goroutines specified to handle updates
// concurrently. This will no longer use inline updates once this function is
// called and all updates will be handled through the goroutines.
func (u *Updater) StartConcurrent(goroutines int) {
	u.isConcurrent = true
	for i := 0; i < goroutines; i++ {
		go u.updateConcurrent()
	}
}

// AddUpdate adds an update function to the list of updates to be called when
// the #Updater.Update function is called. It returns the id of the update
// function that was added so that it can be removed later.
//
// The update function is added to a back-buffer so it will not begin updating
// until the next call to #Updater.Update.
func (u *Updater) AddUpdate(update func(float64)) int {
	id := u.nextId
	u.backAdd = append(u.backAdd, engineUpdate{
		id:     id,
		update: update,
	})
	u.nextId++
	return id
}

// RemoveUpdate removes an update function from the list of updates to be called
// when the #Updater.Update function is called. It takes the id of the update
// function that was returned when the update function was added.
//
// The update function is removed from a back-buffer so it will not be removed
// until the next call to #Updater.Update.
func (u *Updater) RemoveUpdate(id int) {
	if id > 0 {
		u.backRemove = append(u.backRemove, id)
	}
}

// Update calls all of the update functions that have been added to the updater.
// It takes a deltaTime parameter that is the approximate amount of time since
// the last call to #Updater.Update.
func (u *Updater) Update(deltaTime float64) {
	u.lastDelta = deltaTime
	u.addInternal()
	u.removeInternal()
	if u.isConcurrent {
		u.coroutineUpdate()
	} else {
		u.inlineUpdate(deltaTime)
	}
}

// Destroy cleans up the updater and should be called when the updater is no
// longer needed. It will close the pending and complete channels and clear the
// updates map.
func (u *Updater) Destroy() {
	close(u.pending)
	close(u.complete)
	clear(u.updates)
	u.backAdd = u.backAdd[:0]
	u.backRemove = u.backRemove[:0]
}

func (u *Updater) inlineUpdate(deltaTime float64) {
	for i := range u.updates {
		u.updates[i].update(deltaTime)
	}
}

func (u *Updater) coroutineUpdate() {
	waitCount := 0
	for id := range u.updates {
		waitCount++
		u.pending <- id
	}
	for i := 0; i < waitCount; i++ {
		<-u.complete
	}
}

func (u *Updater) updateConcurrent() {
	// TODO:  Does this need to be cleaned up?
	for {
		id := <-u.pending
		u.updates[id].update(u.lastDelta)
		u.complete <- id
	}
}

func (u *Updater) addInternal() {
	for _, update := range u.backAdd {
		u.updates[update.id] = update
	}
	u.backAdd = u.backAdd[:0]
}

func (u *Updater) removeInternal() {
	for _, id := range u.backRemove {
		delete(u.updates, id)
	}
	u.backRemove = u.backRemove[:0]
}
