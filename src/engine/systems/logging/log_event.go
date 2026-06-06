/******************************************************************************/
/* log_event.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package logging

type EventId = int64

type eventEntry struct {
	id   EventId
	call func(string)
}

type Event struct {
	nextId EventId
	calls  []eventEntry
}

func newEvent() Event {
	return Event{
		nextId: 1,
		calls:  make([]eventEntry, 0),
	}
}

func (e Event) IsEmpty() bool { return len(e.calls) == 0 }

func (e *Event) Add(call func(string)) EventId {
	id := e.nextId
	e.nextId++
	e.calls = append(e.calls, eventEntry{id, call})
	return id
}

func (e *Event) Remove(id EventId) {
	for i := range e.calls {
		if e.calls[i].id == id {
			last := len(e.calls) - 1
			e.calls[i], e.calls[last] = e.calls[last], e.calls[i]
			e.calls = e.calls[:last]
			return
		}
	}
}

func (e *Event) Execute(message string) {
	for i := range e.calls {
		e.calls[i].call(message)
	}
}
