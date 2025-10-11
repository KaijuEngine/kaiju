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

import "runtime"

// TODO:  This is a stub and will need some work later

type Threads struct {
	pipe    chan func(threadId int)
	exitSig []chan struct{}
}

func NewThreads() Threads {
	t := Threads{
		pipe: make(chan func(threadId int), 1000),
	}
	return t
}

func (t *Threads) ThreadCount() int { return len(t.exitSig) }

func (t *Threads) Start() {
	t.exitSig = make([]chan struct{}, runtime.NumCPU())
	for i := range len(t.exitSig) {
		t.exitSig[i] = make(chan struct{})
		go t.work(i)
	}
}

func (t *Threads) Stop() {
	for i := range t.exitSig {
		t.exitSig[i] <- struct{}{}
		close(t.exitSig[i])
	}
	close(t.pipe)
}

func (t *Threads) AddWork(work ...func(threadId int)) {
	for i := range work {
		t.pipe <- work[i]
	}
}

func (t *Threads) work(sigIdx int) {
	for len(t.exitSig) > 0 {
		select {
		case <-t.exitSig[sigIdx]:
		case action := <-t.pipe:
			if action != nil {
				action(sigIdx)
			}
		}
	}
}
