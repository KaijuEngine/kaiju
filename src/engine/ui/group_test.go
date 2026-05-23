package ui

import (
	"testing"
	"weak"
)

func TestGroupRejectsInvalidEventType(t *testing.T) {
	t.Parallel()

	group := Group{}
	target := &UI{}
	target.events[EventTypeInvalid].Add(func() {})

	group.requestEvent(target, EventTypeInvalid)

	if len(group.requests) != 0 {
		t.Fatalf("invalid event was queued")
	}
	if group.hadRequests != requestStateNone {
		t.Fatalf("invalid event changed request state to %d", group.hadRequests)
	}
}

func TestGroupLateUpdateClearsUnexpectedInvalidRequest(t *testing.T) {
	t.Parallel()

	man := &Manager{}
	group := Group{
		requests: []groupRequest{{
			target:    &UI{man: weak.Make(man)},
			eventType: EventTypeInvalid,
		}},
	}

	group.lateUpdate()

	if len(group.requests) != 0 {
		t.Fatalf("lateUpdate did not clear invalid request")
	}
	if group.isProcessing {
		t.Fatalf("lateUpdate left group in processing state")
	}
}

func TestGroupLateUpdateDispatchesFocusAndBlur(t *testing.T) {
	t.Parallel()

	man := &Manager{}
	target := &UI{man: weak.Make(man)}
	focused := false
	blurred := false
	target.events[EventTypeFocus].Add(func() { focused = true })
	target.events[EventTypeBlur].Add(func() { blurred = true })
	group := Group{
		requests: []groupRequest{
			{target: target, eventType: EventTypeFocus},
			{target: target, eventType: EventTypeBlur},
		},
	}

	group.lateUpdate()

	if !focused {
		t.Fatalf("focus event was not dispatched")
	}
	if !blurred {
		t.Fatalf("blur event was not dispatched")
	}
	if len(group.requests) != 0 {
		t.Fatalf("lateUpdate did not clear dispatched focus requests")
	}
}
