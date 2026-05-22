/******************************************************************************/
/* named_work.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package concurrent

import (
	"sync"

	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

type WorkGroup struct {
	work  map[string][]func()
	mutex sync.Mutex
}

func (w *WorkGroup) Init() {
	w.work = map[string][]func(){}
}

func (w *WorkGroup) Add(name string, work func()) {
	w.mutex.Lock()
	m := w.work[name]
	m = append(m, work)
	w.work[name] = m
	w.mutex.Unlock()
}

func (w *WorkGroup) Execute(name string, threads *Threads) {
	defer tracing.NewRegion("WorkGroup: " + name).End()
	work := w.work[name]
	if len(work) == 0 {
		return
	}
	calls := make([]func(int), len(work))
	wg := sync.WaitGroup{}
	wg.Add(len(calls))
	for i := range work {
		calls[i] = func(int) {
			work[i]()
			wg.Done()
		}
	}
	threads.AddWork(calls)
	wg.Wait()
	w.work[name] = klib.WipeSlice(work[:0])
}
