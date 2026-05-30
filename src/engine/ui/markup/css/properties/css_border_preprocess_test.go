/******************************************************************************/
/* css_border_preprocess_test.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"testing"

	"kaijuengine.com/engine/ui/markup/css/rules"
)

func testRule(property string, values ...string) rules.Rule {
	out := rules.Rule{Property: property}
	for i := range values {
		out.Values = append(out.Values, rules.PropertyValue{Str: values[i]})
	}
	return out
}

func preprocessBorderTestRules(in []rules.Rule) []rules.Rule {
	all := append([]rules.Rule(nil), in...)
	for i := 0; i < len(all); i++ {
		p, ok := PropertyMap[all[i].Property]
		if !ok {
			continue
		}
		subRules := all[i:]
		all[i].Values, subRules = p.Preprocess(all[i].Values, subRules)
		for j := range subRules {
			all[i+j] = subRules[j]
		}
		all = all[:i+len(subRules)]
	}
	return all
}

func mergedBorderSide(t *testing.T, values []rules.PropertyValue, side int) []rules.PropertyValue {
	t.Helper()
	if len(values) == 0 || values[0].Str != mergedBorderSidesSentinel {
		t.Fatalf("expected merged border values, got %#v", values)
	}
	idx := 1
	for currentSide := 0; currentSide < 4; currentSide++ {
		if idx >= len(values) {
			t.Fatalf("merged border values ended before side %d", currentSide)
		}
		count := int(values[idx].Num)
		idx++
		if idx+count > len(values) {
			t.Fatalf("merged border side %d overruns values %#v", currentSide, values)
		}
		if currentSide == side {
			return values[idx : idx+count]
		}
		idx += count
	}
	t.Fatalf("missing merged border side %d", side)
	return nil
}

func TestBorderPreprocessMergesLaterSideWidthLonghand(t *testing.T) {
	got := preprocessBorderTestRules([]rules.Rule{
		testRule("border", "1px", "solid", "#2A303B"),
		testRule("border-bottom-width", "0px"),
	})

	if len(got) != 1 {
		t.Fatalf("expected one merged border rule, got %#v", got)
	}
	side := mergedBorderSide(t, got[0].Values, borderSideBottom)
	if side[0].Str != "0px" || side[1].Str != "solid" {
		t.Fatalf("expected bottom width override while preserving style, got %#v", side)
	}
}

func TestBorderSidePreprocessMergesLaterSideWidthLonghand(t *testing.T) {
	got := preprocessBorderTestRules([]rules.Rule{
		testRule("border-bottom", "1px", "solid", "#2A303B"),
		testRule("border-bottom-width", "0px"),
	})

	if len(got) != 1 {
		t.Fatalf("expected one border-bottom rule, got %#v", got)
	}
	if got[0].Property != "border-bottom" || got[0].Values[0].Str != "0px" || got[0].Values[1].Str != "solid" {
		t.Fatalf("expected merged border-bottom width override, got %#v", got)
	}
}

func TestBorderSideWidthPreprocessDropsWhenLaterSideShorthandOverrides(t *testing.T) {
	got := preprocessBorderTestRules([]rules.Rule{
		testRule("border-bottom-width", "0px"),
		testRule("border-bottom", "1px", "solid", "#2A303B"),
	})

	if len(got) != 1 {
		t.Fatalf("expected width longhand to be removed, got %#v", got)
	}
	if got[0].Property != "border-bottom" || got[0].Values[0].Str != "1px" {
		t.Fatalf("expected later border-bottom shorthand to remain, got %#v", got)
	}
}

func TestBorderWidthPreprocessMatchesLaterSideShorthandWidth(t *testing.T) {
	got := preprocessBorderTestRules([]rules.Rule{
		testRule("border-width", "0px"),
		testRule("border-bottom", "1px", "solid", "#2A303B"),
	})

	if len(got) != 2 {
		t.Fatalf("expected border-width and border-bottom rules, got %#v", got)
	}
	if got[0].Property != "border-width" || got[0].Values[2].Str != "1px" {
		t.Fatalf("expected border-width bottom value to match later side shorthand, got %#v", got[0])
	}
	if got[1].Property != "border-bottom" {
		t.Fatalf("expected later border-bottom to remain for style/color, got %#v", got[1])
	}
}
