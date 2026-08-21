/******************************************************************************/
/* html_element_layout_stylizer_test.go                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package document

import (
	"testing"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
)

type styleDiffTestProperty struct {
	key    string
	impact StyleImpact
}

func (p styleDiffTestProperty) Key() string { return p.key }
func (p styleDiffTestProperty) Process(*ui.Panel, *Element, []rules.PropertyValue, *engine.Host) error {
	return nil
}
func (p styleDiffTestProperty) Sort() int { return 0 }
func (p styleDiffTestProperty) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	return values, ruleList
}
func (p styleDiffTestProperty) StyleImpact() StyleImpact { return p.impact }

func withStyleDiffProperties(t *testing.T) {
	t.Helper()
	previous := LinkedPropertyMap
	LinkedPropertyMap = map[string]CSSProperty{
		"width":        styleDiffTestProperty{key: "width", impact: StyleImpactLayout},
		"border":       styleDiffTestProperty{key: "border", impact: StyleImpactLayout},
		"border-color": styleDiffTestProperty{key: "border-color", impact: StyleImpactPaint},
		"color":        styleDiffTestProperty{key: "color", impact: StyleImpactPaint},
	}
	t.Cleanup(func() { LinkedPropertyMap = previous })
}

func testComputedRule(property string, values ...string) rules.Rule {
	r := rules.Rule{Property: property, Values: make([]rules.PropertyValue, len(values))}
	for i := range values {
		r.Values[i].Str = values[i]
	}
	return r
}

func TestComputedRuleDiffIgnoresIdenticalLayoutRules(t *testing.T) {
	withStyleDiffProperties(t)
	previous := []rules.Rule{
		testComputedRule("width", "100px"),
		testComputedRule("border-color", "red"),
	}
	next := []rules.Rule{
		testComputedRule("width", "100px"),
		testComputedRule("border-color", "blue"),
	}

	layoutChanged, paintChanged := computedRuleDiff(previous, next)
	if layoutChanged {
		t.Fatal("color-only computed change should not affect layout")
	}
	if !paintChanged {
		t.Fatal("color-only computed change should affect paint")
	}
}

func TestComputedRuleDiffDetectsLayoutValueChange(t *testing.T) {
	withStyleDiffProperties(t)
	previous := []rules.Rule{testComputedRule("width", "100px")}
	next := []rules.Rule{testComputedRule("width", "120px")}

	layoutChanged, _ := computedRuleDiff(previous, next)
	if !layoutChanged {
		t.Fatal("changed computed width should affect layout")
	}
}

func TestComputedRuleDiffTreatsBorderShorthandColorAsPaint(t *testing.T) {
	withStyleDiffProperties(t)
	previous := []rules.Rule{testComputedRule("border", "1px", "solid", "red")}
	next := []rules.Rule{testComputedRule("border", "1px", "solid", "blue")}

	layoutChanged, paintChanged := computedRuleDiff(previous, next)
	if layoutChanged {
		t.Fatal("border shorthand color-only change should not affect layout")
	}
	if !paintChanged {
		t.Fatal("border shorthand color-only change should affect paint")
	}
}

func TestComputedRuleDiffIdenticalRulesDoNothing(t *testing.T) {
	withStyleDiffProperties(t)
	computed := []rules.Rule{
		testComputedRule("width", "100px"),
		testComputedRule("color", "red"),
	}

	layoutChanged, paintChanged := computedRuleDiff(computed, rules.CloneRules(computed))
	if layoutChanged || paintChanged {
		t.Fatalf("identical computed rules changed layout=%t paint=%t", layoutChanged, paintChanged)
	}
}
