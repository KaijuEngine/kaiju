/******************************************************************************/
/* content_loader.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package content_workspace

import "kaijuengine.com/engine"

const DefaultContentLoadBatchSize = 48

type BatchedContentLoader struct {
	host       *engine.Host
	batchSize  int
	pending    []string
	queued     map[string]struct{}
	running    bool
	generation int
	process    func([]string)
}

func (l *BatchedContentLoader) Configure(host *engine.Host, batchSize int, process func([]string)) {
	l.host = host
	l.batchSize = batchSize
	l.process = process
	if l.batchSize <= 0 {
		l.batchSize = DefaultContentLoadBatchSize
	}
	if l.queued == nil {
		l.queued = make(map[string]struct{})
	}
}

func (l *BatchedContentLoader) Enqueue(ids []string) {
	if len(ids) == 0 || l.process == nil {
		return
	}
	if l.host == nil {
		l.process(ids)
		return
	}
	if l.queued == nil {
		l.queued = make(map[string]struct{})
	}
	for i := range ids {
		if _, ok := l.queued[ids[i]]; ok {
			continue
		}
		l.queued[ids[i]] = struct{}{}
		l.pending = append(l.pending, ids[i])
	}
	l.schedule()
}

func (l *BatchedContentLoader) Stop() {
	l.pending = l.pending[:0]
	clear(l.queued)
	l.running = false
	l.generation++
}

func (l *BatchedContentLoader) schedule() {
	if l.running || len(l.pending) == 0 {
		return
	}
	l.running = true
	generation := l.generation
	l.host.RunAfterFrames(0, func() {
		l.run(generation)
	})
}

func (l *BatchedContentLoader) run(generation int) {
	if generation != l.generation {
		return
	}
	if len(l.pending) == 0 {
		l.running = false
		return
	}
	count := min(l.batchSize, len(l.pending))
	batch := append([]string(nil), l.pending[:count]...)
	for i := range batch {
		delete(l.queued, batch[i])
	}
	copy(l.pending, l.pending[count:])
	l.pending = l.pending[:len(l.pending)-count]
	l.process(batch)
	if len(l.pending) == 0 {
		l.running = false
		return
	}
	l.host.RunAfterFrames(0, func() {
		l.run(generation)
	})
}
