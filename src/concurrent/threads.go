package concurrent

import "runtime"

// TODO:  This is a stub and will need some work later

type Threads struct {
	pipe    chan func()
	exitSig []chan struct{}
}

func NewThreads() Threads {
	t := Threads{
		pipe: make(chan func(), 1000),
	}
	return t
}

func (t *Threads) Start() {
	t.exitSig = make([]chan struct{}, runtime.NumCPU())
	for i := range len(t.exitSig) {
		t.exitSig[i] = make(chan struct{}, 0)
		go t.work(i)
	}
}

func (t *Threads) Stop() {
	for i := range t.exitSig {
		t.exitSig[i] <- struct{}{}
	}
}

func (t *Threads) AddWork(work ...func()) {
	for i := range work {
		t.pipe <- work[i]
	}
}

func (t *Threads) work(sigIdx int) {
	for {
		select {
		case <-t.exitSig[sigIdx]:
			break
		case action := <-t.pipe:
			action()
		}
	}
}
