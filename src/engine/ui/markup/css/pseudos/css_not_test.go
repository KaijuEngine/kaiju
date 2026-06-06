/******************************************************************************/
/* css_not_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"testing"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func testNotElement(t *testing.T) *document.Element {
	t.Helper()
	root := document.NewHTML(`<button id="save" class="primary materialIcon" type="button"></button>`)
	elm := root.FindElementByTag("button")
	if elm == nil {
		t.Fatal("failed to create test element")
	}
	return elm
}

func TestNotRejectsMatchingSelectors(t *testing.T) {
	elm := testNotElement(t)
	tests := [][]string{
		{".", "primary"},
		{"#", "save"},
		{"button"},
		{"[", "type", "=", "button", "]"},
	}
	for _, args := range tests {
		got, err := (Not{}).Process(elm, rules.SelectorPart{Args: args})
		if err != nil {
			t.Fatalf(":not(%v) returned an error: %v", args, err)
		}
		if len(got) != 0 {
			t.Fatalf(":not(%v) should reject a matching element", args)
		}
	}
}

func TestNotKeepsNonMatchingSelectors(t *testing.T) {
	elm := testNotElement(t)
	tests := [][]string{
		{".", "secondary"},
		{"#", "cancel"},
		{"span"},
		{"[", "type", "=", "submit", "]"},
		{".", "secondary", ",", "#", "cancel"},
	}
	for _, args := range tests {
		got, err := (Not{}).Process(elm, rules.SelectorPart{Args: args})
		if err != nil {
			t.Fatalf(":not(%v) returned an error: %v", args, err)
		}
		if len(got) != 1 || got[0] != elm {
			t.Fatalf(":not(%v) should keep a non-matching element", args)
		}
	}
}

func TestNotRejectsIfAnySelectorInListMatches(t *testing.T) {
	elm := testNotElement(t)
	got, err := (Not{}).Process(elm, rules.SelectorPart{
		Args: []string{".", "secondary", ",", ".", "primary"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatal(":not should reject when any selector in the list matches")
	}
}
