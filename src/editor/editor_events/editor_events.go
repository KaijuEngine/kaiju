package editor_events

import "kaiju/engine/systems/events"

type EditorEvents struct {
	// OnContentAdded sends list of content ids that have been added
	OnContentAdded events.EventWithArg[[]string]
}
