package concurrent

import (
	"kaiju/profiler/tracing"
	"sync"
)

type WorkGroup struct {
	work sync.Map
}

func (w *WorkGroup) Add(name string, work func()) {
	if target, loaded := w.work.LoadOrStore(name, []func(){work}); loaded {
		list := target.([]func())
		list = append(list, work)
		w.work.Store(name, list)
	}
}

func (w *WorkGroup) Execute(name string, threads *Threads) {
	defer tracing.NewRegion("WorkGroup: " + name).End()
	if target, ok := w.work.Load(name); ok {
		list := target.([]func())
		wg := sync.WaitGroup{}
		wg.Add(len(list))
		for i := range list {
			threads.AddWork(func() {
				list[i]()
				wg.Done()
			})
		}
		wg.Wait()
		list = list[:0]
		w.work.Store(name, list)
	}
}
