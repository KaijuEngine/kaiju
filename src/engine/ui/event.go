/******************************************************************************/
/* event.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

type EventType = int

const (
	EventTypeInvalid = EventType(iota)
	EventTypeEnter
	EventTypeMove
	EventTypeExit
	EventTypeClick
	EventTypeRightClick
	EventTypeDoubleClick
	EventTypeDown
	EventTypeUp
	EventTypeRightDown
	EventTypeRightUp
	EventTypeMiss
	EventTypeDragStart
	EventTypeDrop
	EventTypeDropEnter
	EventTypeDropExit
	EventTypeDragEnd
	EventTypeScroll
	EventTypeRebuild
	EventTypeDestroy
	EventTypeFocus
	EventTypeBlur
	EventTypeSubmit
	EventTypeChange
	EventTypeRender
	EventTypeKeyDown
	EventTypeKeyUp
	EventTypeEnd
)
