/******************************************************************************/
/* parser_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package css

import (
	"testing"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
)

func TestParser(t *testing.T) {
	// s := rules.NewStyleSheet()
	// s.Parse(DefaultCSS)
	// if len(s.Groups) == 0 {
	// 	t.Error("No groups found")
	// }
}

func TestCSSMapAddRemovesEarlierLonghandWhenShorthandOverrides(t *testing.T) {
	elm := &ui.UI{}
	cssMap := CSSMap{}

	cssMap.add(elm, []rules.Rule{{Property: "margin-bottom"}})
	cssMap.add(elm, []rules.Rule{{Property: "margin"}})

	got := cssMap[elm]
	if len(got) != 1 {
		t.Fatalf("expected one rule after shorthand override, got %d", len(got))
	}
	if got[0].Property != "margin" {
		t.Fatalf("expected margin to remain, got %q", got[0].Property)
	}
}

func TestCSSMapAddKeepsShorthandWhenLaterLonghandOverridesOneSide(t *testing.T) {
	elm := &ui.UI{}
	cssMap := CSSMap{}

	cssMap.add(elm, []rules.Rule{{Property: "margin"}})
	cssMap.add(elm, []rules.Rule{{Property: "margin-bottom"}})

	got := cssMap[elm]
	if len(got) != 2 {
		t.Fatalf("expected shorthand and later longhand to remain, got %d", len(got))
	}
	if got[0].Property != "margin" || got[1].Property != "margin-bottom" {
		t.Fatalf("expected margin then margin-bottom, got %#v", got)
	}
}

func TestCleanMapDuplicatesUnderstandsShorthandOverrides(t *testing.T) {
	elm := &ui.UI{}
	cssMap := CSSMap{
		elm: {
			{Property: "padding-top"},
			{Property: "padding-right"},
			{Property: "padding"},
		},
	}

	cleanMapDuplicates(cssMap)

	got := cssMap[elm]
	if len(got) != 1 {
		t.Fatalf("expected one rule after duplicate cleanup, got %d", len(got))
	}
	if got[0].Property != "padding" {
		t.Fatalf("expected padding to remain, got %q", got[0].Property)
	}
}
