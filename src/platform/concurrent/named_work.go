/******************************************************************************/
/* named_work.go                                                              */
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
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
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
			threads.AddWork(func(int) {
				list[i]()
				wg.Done()
			})
		}
		wg.Wait()
		list = klib.WipeSlice(list[:0])
		w.work.Store(name, list)
	}
}
