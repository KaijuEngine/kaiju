/******************************************************************************/
/* traced_log_event.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package logging

type tracedEventEntry struct {
	id   EventId
	call func(string, []string)
}

type TracedEvent struct {
	nextId EventId
	calls  []tracedEventEntry
}

func newTracedEvent() TracedEvent {
	return TracedEvent{
		nextId: 1,
		calls:  make([]tracedEventEntry, 0),
	}
}

func (e TracedEvent) IsEmpty() bool { return len(e.calls) == 0 }

func (e *TracedEvent) Add(call func(msg string, trace []string)) EventId {
	id := e.nextId
	e.nextId++
	e.calls = append(e.calls, tracedEventEntry{id, call})
	return id
}

func (e *TracedEvent) Remove(id EventId) {
	for i := range e.calls {
		if e.calls[i].id == id {
			last := len(e.calls) - 1
			e.calls[i], e.calls[last] = e.calls[last], e.calls[i]
			e.calls = e.calls[:last]
			return
		}
	}
}

func (e *TracedEvent) Execute(message string, trace []string) {
	for i := range e.calls {
		e.calls[i].call(message, trace)
	}
}
