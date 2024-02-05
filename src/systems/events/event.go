package events

type Id = int64

type eventEntry struct {
	id   Id
	call func()
}

type Event struct {
	nextId Id
	calls  []eventEntry
}

func New() Event {
	return Event{
		nextId: 1,
		calls:  make([]eventEntry, 0),
	}
}

func (e Event) IsEmpty() bool { return len(e.calls) == 0 }

func (e *Event) Add(call func()) Id {
	id := e.nextId
	e.nextId++
	e.calls = append(e.calls, eventEntry{id, call})
	return id
}

func (e *Event) Remove(id Id) {
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
	for i := range e.calls {
		e.calls[i].call()
	}
}
