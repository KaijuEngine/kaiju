/******************************************************************************/
/* ui_disabled_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package ui

import "testing"

func TestDisabledSuppressesRequestedEventAndBlocksPointer(t *testing.T) {
	t.Parallel()

	target := &UI{}
	target.SetDisabled(true)

	clicked := false
	target.AddEvent(EventTypeClick, func() { clicked = true })

	if !target.requestEvent(EventTypeClick) {
		t.Fatal("disabled pointer event should still block sibling hit targets")
	}
	if clicked {
		t.Fatal("disabled requested event should not execute callbacks")
	}
}

func TestDisabledAllowsProgrammaticExecuteEvent(t *testing.T) {
	t.Parallel()

	target := &UI{}
	target.SetDisabled(true)

	clicked := false
	target.AddEvent(EventTypeClick, func() { clicked = true })

	if !target.ExecuteEvent(EventTypeClick) {
		t.Fatal("programmatic ExecuteEvent should report registered callbacks")
	}
	if !clicked {
		t.Fatal("programmatic ExecuteEvent should execute callbacks while disabled")
	}
}
