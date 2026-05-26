/******************************************************************************/
/* event_with_arg.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package events

import "kaijuengine.com/platform/profiler/tracing"

type eventWithArgEntry[T any] struct {
	id   Id
	call func(arg T)
}

type EventWithArg[T any] struct {
	nextId Id
	calls  []eventWithArgEntry[T]
}

func (e EventWithArg[T]) IsEmpty() bool { return len(e.calls) == 0 }

func (e *EventWithArg[T]) Add(call func(arg T)) Id {
	e.nextId++
	id := e.nextId
	e.calls = append(e.calls, eventWithArgEntry[T]{id, call})
	return id
}

func (e *EventWithArg[T]) Clear() {
	e.calls = e.calls[:0]
	e.nextId = 0
}

func (e *EventWithArg[T]) Remove(id Id) {
	defer tracing.NewRegion("Event.Remove").End()
	for i := range e.calls {
		if e.calls[i].id == id {
			last := len(e.calls) - 1
			e.calls[i], e.calls[last] = e.calls[last], e.calls[i]
			e.calls = e.calls[:last]
			return
		}
	}
}

func (e *EventWithArg[T]) Execute(arg T) {
	defer tracing.NewRegion("Event.Execute").End()
	for i := range e.calls {
		e.calls[i].call(arg)
	}
}
