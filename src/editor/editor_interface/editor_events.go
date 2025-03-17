package editor_interface

import "kaiju/engine/systems/events"

type ContentSelect struct {
	Event   events.Event
	Content []string
}

type EditorEvents struct {
	OnContentSelect ContentSelect
}
