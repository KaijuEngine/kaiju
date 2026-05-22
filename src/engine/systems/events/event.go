/******************************************************************************/
/* event.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package events

import (
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

type Id = int64

type eventEntry struct {
	id   Id
	call func()
}

type Event struct {
	nextId Id
	calls  []eventEntry
}

func (e Event) IsEmpty() bool { return len(e.calls) == 0 }

func (e *Event) Add(call func()) Id {
	e.nextId++
	id := e.nextId
	e.calls = append(e.calls, eventEntry{id, call})
	return id
}

func (e *Event) Clear() {
	e.calls = klib.WipeSlice(e.calls)
	e.nextId = 0
}

func (e *Event) Remove(id Id) {
	defer tracing.NewRegion("Event.Remove").End()
	if id == 0 {
		return
	}
	for i := range e.calls {
		if e.calls[i].id == id {
			last := len(e.calls) - 1
			e.calls[i], e.calls[last] = e.calls[last], e.calls[i]
			e.calls = e.calls[:last]
			return
		}
	}
}

func (e *Event) Execute() {
	defer tracing.NewRegion("Event.Execute").End()
	for i := range e.calls {
		e.calls[i].call()
	}
}
