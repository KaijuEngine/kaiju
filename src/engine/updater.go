/*****************************************************************************/
/* updater.go                                                                */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
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
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package engine

type engineUpdate struct {
	id     int
	update func(float64)
}

type Updater struct {
	updates    map[int]engineUpdate
	backAdd    []engineUpdate
	backRemove []int
	nextId     int
	lastDelta  float64
	pending    chan int
	complete   chan int
}

func NewUpdater() Updater {
	return Updater{
		updates:    make(map[int]engineUpdate),
		backAdd:    make([]engineUpdate, 0),
		backRemove: make([]int, 0),
		nextId:     1,
		pending:    make(chan int, 100),
		complete:   make(chan int, 100),
	}
}

func (u *Updater) StartThreads(threads int) {
	for i := 0; i < threads; i++ {
		go u.updateThread()
	}
}

func (u *Updater) updateThread() {
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

func (u *Updater) AddUpdate(update func(float64)) int {
	id := u.nextId
	u.backAdd = append(u.backAdd, engineUpdate{
		id:     id,
		update: update,
	})
	u.nextId++
	return id
}

func (u *Updater) RemoveUpdate(id int) {
	if id > 0 {
		u.backRemove = append(u.backRemove, id)
	}
}

func (u *Updater) inlineUpdate(deltaTime float64) {
	for i := range u.updates {
		u.updates[i].update(deltaTime)
	}
}

func (u *Updater) threadedUpdate() {
	waitCount := 0
	for id := range u.updates {
		waitCount++
		u.pending <- id
	}
	for i := 0; i < waitCount; i++ {
		<-u.complete
	}
}

func (u *Updater) Update(deltaTime float64) {
	u.lastDelta = deltaTime
	u.addInternal()
	u.removeInternal()
	u.inlineUpdate(deltaTime)
	//u.threadedUpdate()
}

func (u *Updater) Destroy() {
	close(u.pending)
	close(u.complete)
	clear(u.updates)
	u.backAdd = u.backAdd[:0]
	u.backRemove = u.backRemove[:0]
}
