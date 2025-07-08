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
	}
}

func (t *Threads) AddWork(work ...func(threadId int)) {
	for i := range work {
		t.pipe <- work[i]
	}
}

func (t *Threads) work(sigIdx int) {
	for {
		select {
		case <-t.exitSig[sigIdx]:
		case action := <-t.pipe:
			action(sigIdx)
		}
	}
}
