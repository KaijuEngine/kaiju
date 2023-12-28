package ui

import "slices"

type EventType = int

const (
	EventTypeInvalid = iota
	EventTypeEnter
	EventTypeExit
	EventTypeClick
	EventTypeDown
	EventTypeUp
	EventTypeMiss
	EventTypeScroll
	EventTypeRebuild
	EventTypeDestroy
	EventTypeChange
	EventTypeEnd
)

type EventId = int

type uiEventEntry struct {
	id   EventId
	call func()
}

type uiEvent struct {
	nextId   EventId
	uiEvents []uiEventEntry
}

func (evt *uiEvent) isEmpty() bool { return len(evt.uiEvents) == 0 }

func (evt *uiEvent) add(call func()) EventId {
	id := evt.nextId
	evt.nextId++
	evt.uiEvents = append(evt.uiEvents, uiEventEntry{id: id, call: call})
	return id
}

func (evt *uiEvent) remove(id EventId) {
	for i := 0; i < len(evt.uiEvents); i++ {
		if evt.uiEvents[i].id == id {
			evt.uiEvents = slices.Delete(evt.uiEvents, i, i+1)
			return
		}
	}
}

func (evt *uiEvent) execute() {
	for _, entry := range evt.uiEvents {
		entry.call()
	}
}
