/******************************************************************************/
/* threads.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package concurrent

import (
	"container/list"
	"runtime"
	"sync"

	"kaijuengine.com/platform/profiler/tracing"
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
	t.cond.Broadcast()
	for t.count > 0 {
		t.cond.Wait()
	}
	t.mutex.Unlock()
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
	defer func() {
		t.mutex.Lock()
		t.count--
		t.cond.Broadcast()
		t.mutex.Unlock()
	}()
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
