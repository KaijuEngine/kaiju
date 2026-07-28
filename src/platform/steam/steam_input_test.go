/******************************************************************************/
/* steam_input_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package steam

import "testing"

func TestDigitalActionTransitions(t *testing.T) {
	data := nextDigitalActionData(DigitalActionData{}, false, true)
	if data.Pressed || data.Held || data.Released {
		t.Fatalf("idle action produced an edge: %+v", data)
	}

	data = nextDigitalActionData(data, true, true)
	if !data.Pressed || data.Held || data.Released {
		t.Fatalf("initial down state = %+v, want pressed", data)
	}

	data = nextDigitalActionData(data, true, true)
	if data.Pressed || !data.Held || data.Released {
		t.Fatalf("continued down state = %+v, want held", data)
	}

	data = nextDigitalActionData(data, false, true)
	if data.Pressed || data.Held || !data.Released {
		t.Fatalf("up state = %+v, want released", data)
	}

	data = nextDigitalActionData(data, false, true)
	if data.Pressed || data.Held || data.Released {
		t.Fatalf("continued up state produced an edge: %+v", data)
	}
}

func TestDigitalActionBecomingInactiveReleases(t *testing.T) {
	previous := DigitalActionData{State: true, Active: true, Held: true}
	data := nextDigitalActionData(previous, true, false)
	if data.Pressed || data.Held || !data.Released {
		t.Fatalf("inactive action state = %+v, want released", data)
	}
}

func TestControllersReturnsCopy(t *testing.T) {
	system := newInputSystem()
	system.controllers = []InputHandle{17}

	controllers := system.Controllers()
	controllers[0] = 42

	if system.controllers[0] != 17 {
		t.Fatalf("Controllers exposed internal storage: %v", system.controllers)
	}
}
