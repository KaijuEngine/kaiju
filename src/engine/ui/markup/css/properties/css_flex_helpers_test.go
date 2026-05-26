/******************************************************************************/
/* css_flex_helpers_test.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"testing"

	"kaijuengine.com/engine/ui"
)

func TestParseFlexFloat(t *testing.T) {
	if got, ok := parseFlexFloat("2.5"); !ok || got != 2.5 {
		t.Fatalf("expected 2.5 to parse, got %.2f ok=%v", got, ok)
	}
	if _, ok := parseFlexFloat("auto"); ok {
		t.Fatal("expected auto not to parse as a flex number")
	}
}

func TestParseFlexAlign(t *testing.T) {
	tests := map[string]ui.FlexAlign{
		"auto":       ui.FlexAlignAuto,
		"stretch":    ui.FlexAlignStretch,
		"flex-start": ui.FlexAlignStart,
		"flex-end":   ui.FlexAlignEnd,
		"center":     ui.FlexAlignCenter,
		"baseline":   ui.FlexAlignStart,
	}
	for value, want := range tests {
		got, ok := parseFlexAlign(value)
		if !ok {
			t.Fatalf("expected %q to parse", value)
		}
		if got != want {
			t.Fatalf("expected %q to parse as %d, got %d", value, want, got)
		}
	}
	if _, ok := parseFlexAlign("banana"); ok {
		t.Fatal("expected invalid alignment not to parse")
	}
}
