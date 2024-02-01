package ui

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
