/******************************************************************************/
/* threads.go                                                                 */
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

package concurrent

import (
	"container/list"
	"kaiju/platform/profiler/tracing"
	"runtime"
	"sync"
)

type Threads struct {
	queue    *list.List
	mutex    sync.Mutex
	cond     *sync.Cond
	shutdown bool
	count    int
}

func (t *Threads) Initialize() {
	defer tracing.NewRegion("Threads.Initialize").End()
	t.queue = list.New()
	t.cond = sync.NewCond(&t.mutex)
}

func (t *Threads) ThreadCount() int { return t.count }

func (t *Threads) Start() {
	defer tracing.NewRegion("Threads.Start").End()
	t.count = runtime.NumCPU()
	for i := 0; i < t.count; i++ {
		go t.work(i)
	}
}

func (t *Threads) Stop() {
	defer tracing.NewRegion("Threads.Stop").End()
	t.mutex.Lock()
	t.shutdown = true
	t.mutex.Unlock()
	t.cond.Broadcast()
}

func (t *Threads) AddWork(work []func(threadId int)) {
	defer tracing.NewRegion("Threads.AddWork").End()
	if len(work) == 0 {
		return
	}
	t.mutex.Lock()
	for _, w := range work {
		t.queue.PushBack(w)
	}
	t.mutex.Unlock()
	t.cond.Broadcast()
}

func (t *Threads) work(id int) {
	for {
		t.mutex.Lock()
		if t.shutdown {
			t.mutex.Unlock()
			return
		}
		for t.queue != nil && t.queue.Len() == 0 {
			t.cond.Wait()
			if t.shutdown {
				t.mutex.Unlock()
				return
			}
		}
		if t.queue == nil {
			t.mutex.Unlock()
			return
		}
		elem := t.queue.Front()
		t.queue.Remove(elem)
		t.mutex.Unlock()
		action := elem.Value.(func(int))
		if action != nil {
			action(id)
		}
	}
}
