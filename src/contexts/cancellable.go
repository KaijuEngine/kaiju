package contexts

import (
	"context"
	"time"
)

type Cancellable struct {
	cancelled chan struct{}
	done      bool
}

func NewCancellable() *Cancellable {
	return &Cancellable{
		cancelled: make(chan struct{}),
	}
}

func (p *Cancellable) Cancel() {
	p.cancelled <- struct{}{}
	close(p.cancelled)
	p.done = true
}

func (p *Cancellable) Deadline() (time.Time, bool) { return time.Time{}, false }
func (p *Cancellable) Done() <-chan struct{}       { return p.cancelled }
func (p *Cancellable) Value(any) any               { return nil }

func (p *Cancellable) Err() error {
	if p.done {
		return context.Canceled
	}
	return nil
}
