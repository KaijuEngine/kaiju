/******************************************************************************/
/* parser_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rules

import "testing"

const testCSSNarrowTag = `.entry span { display: none; }`
const testCSSNarrowClass = `.entry .wide { display: none; }`
const testCSSChildClass = `.entry > .direct { display: none; }`
const testCSSCommaId = `#id1, #id2, #id3 { display: none; }`
const testCSSInputText = `input[type="text"] { display: none; }`
const testCSSNotClass = `button:not(.materialIcon) { display: none; }`
const testCSSNotId = `.idTile:not(#idBlocked) { display: none; }`
const testCSSVarDeclare = `:root { --ed-menu-bar-height: 24px; }
.test { height: var(--ed-menu-bar-height); }`
const testCSSVarInCalc = `:root { --ed-menu-bar-height: 24px; }
.test { height: calc(100% - var(--ed-menu-bar-height)); }`

type dummyWindow struct{}

func (dummyWindow) DotsPerMillimeter() float64 { return 1 }
func (dummyWindow) Width() int                 { return 0 }
func (dummyWindow) Height() int                { return 0 }

func TestParseNarrowTag(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSNarrowTag, dummyWindow{})
	if len(s.Groups) != 1 {
		t.FailNow()
	}
	g := s.Groups[0]
	if len(g.Selectors) != 1 {
		t.FailNow()
	}
	sel := g.Selectors[0]
	if len(sel.Parts) != 3 {
		t.FailNow()
	}
	if sel.Parts[0].Name != "entry" {
		t.FailNow()
	}
	if sel.Parts[0].SelectType != ReadingClass {
		t.FailNow()
	}
	if sel.Parts[1].Name != " " {
		t.FailNow()
	}
	if sel.Parts[1].SelectType != ReadingDescendant {
		t.FailNow()
	}
	if sel.Parts[2].Name != "span" {
		t.FailNow()
	}
	if sel.Parts[2].SelectType != ReadingTag {
		t.FailNow()
	}
}

func TestParseNarrowClass(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSNarrowClass, dummyWindow{})
	if len(s.Groups) != 1 {
		t.FailNow()
	}
	g := s.Groups[0]
	if len(g.Selectors) != 1 {
		t.FailNow()
	}
	sel := g.Selectors[0]
	if len(sel.Parts) != 3 {
		t.FailNow()
	}
	if sel.Parts[0].Name != "entry" {
		t.FailNow()
	}
	if sel.Parts[0].SelectType != ReadingClass {
		t.FailNow()
	}
	if sel.Parts[1].Name != " " {
		t.FailNow()
	}
	if sel.Parts[1].SelectType != ReadingDescendant {
		t.FailNow()
	}
	if sel.Parts[2].Name != "wide" {
		t.FailNow()
	}
	if sel.Parts[2].SelectType != ReadingClass {
		t.FailNow()
	}
}

func TestParseChildClass(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSChildClass, dummyWindow{})
	if len(s.Groups) != 1 {
		t.FailNow()
	}
	g := s.Groups[0]
	if len(g.Selectors) != 1 {
		t.FailNow()
	}
	sel := g.Selectors[0]
	if len(sel.Parts) != 3 {
		t.Fatalf("expected 3 selector parts, got %d: %#v", len(sel.Parts), sel.Parts)
	}
	if sel.Parts[0].Name != "entry" || sel.Parts[0].SelectType != ReadingClass {
		t.FailNow()
	}
	if sel.Parts[1].Name != ">" || sel.Parts[1].SelectType != ReadingChild {
		t.Fatalf("expected child combinator, got %#v", sel.Parts[1])
	}
	if sel.Parts[2].Name != "direct" || sel.Parts[2].SelectType != ReadingClass {
		t.FailNow()
	}
}

func TestParseCommaIds(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSCommaId, dummyWindow{})
	if len(s.Groups) != 3 {
		t.FailNow()
	}
	for i := range s.Groups {
		if len(s.Groups[i].Rules) != 1 {
			t.FailNow()
		}
		if s.Groups[i].Rules[0].Property != "display" {
			t.FailNow()
		}
		if len(s.Groups[i].Rules[0].Values) != 1 {
			t.FailNow()
		}
		if s.Groups[i].Rules[0].Values[0].Str != "none" {
			t.FailNow()
		}
	}
}

func TestParseTextSubType(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSInputText, dummyWindow{})
	if len(s.Groups) != 1 {
		t.FailNow()
	}
	if len(s.Groups[0].Selectors) != 1 {
		t.FailNow()
	}
	if len(s.Groups[0].Selectors[0].Parts) != 3 {
		t.FailNow()
	}
	p := s.Groups[0].Selectors[0].Parts
	if p[0].Name != "input" {
		t.FailNow()
	}
	if p[1].Name != "type" {
		t.FailNow()
	}
	if p[2].Name != "text" {
		t.FailNow()
	}
}

func TestParseNotFunction(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSNotClass, dummyWindow{})
	if len(s.Groups) != 1 {
		t.FailNow()
	}
	if len(s.Groups[0].Selectors) != 1 {
		t.FailNow()
	}
	p := s.Groups[0].Selectors[0].Parts
	if len(p) != 2 {
		t.Fatalf("expected selector to have 2 parts, got %d", len(p))
	}
	if p[0].Name != "button" || p[0].SelectType != ReadingTag {
		t.FailNow()
	}
	if p[1].Name != "not" || p[1].SelectType != ReadingPseudoFunction {
		t.FailNow()
	}
	expectedArgs := []string{".", "materialIcon"}
	if len(p[1].Args) != len(expectedArgs) {
		t.Fatalf("expected %d :not args, got %d: %#v", len(expectedArgs), len(p[1].Args), p[1].Args)
	}
	for i := range expectedArgs {
		if p[1].Args[i] != expectedArgs[i] {
			t.Fatalf("expected arg %d to be %q, got %q", i, expectedArgs[i], p[1].Args[i])
		}
	}
}

func TestParseNotFunctionId(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSNotId, dummyWindow{})
	if len(s.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d: %#v", len(s.Groups), s.Groups)
	}
	if len(s.Groups[0].Selectors) != 1 {
		t.Fatalf("expected 1 selector, got %d: %#v", len(s.Groups[0].Selectors), s.Groups[0].Selectors)
	}
	p := s.Groups[0].Selectors[0].Parts
	if len(p) != 2 {
		t.Fatalf("expected selector to have 2 parts, got %d", len(p))
	}
	expectedArgs := []string{"#idBlocked"}
	if len(p[1].Args) != len(expectedArgs) {
		t.Fatalf("expected %d :not args, got %d: %#v", len(expectedArgs), len(p[1].Args), p[1].Args)
	}
	for i := range expectedArgs {
		if p[1].Args[i] != expectedArgs[i] {
			t.Fatalf("expected arg %d to be %q, got %q in %#v", i, expectedArgs[i], p[1].Args[i], p[1].Args)
		}
	}
}

func TestParseVariable(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSVarDeclare, dummyWindow{})
	if len(s.Groups) != 2 {
		t.FailNow()
	}
	if len(s.CustomVars) != 1 {
		t.FailNow()
	}
	if v, ok := s.CustomVars["--ed-menu-bar-height"]; !ok {
		t.FailNow()
	} else if len(v) != 1 {
		t.FailNow()
	} else if v[0] != "24px" {
		t.FailNow()
	}
	if len(s.Groups[1].Rules) != 1 {
		t.FailNow()
	}
	if s.Groups[1].Rules[0].Property != "height" {
		t.FailNow()
	}
	if len(s.Groups[1].Rules[0].Values) != 1 {
		t.FailNow()
	}
	if s.Groups[1].Rules[0].Values[0].Str != "24px" {
		t.FailNow()
	}
}

func TestParseCalcAndVariable(t *testing.T) {
	s := NewStyleSheet()
	s.Parse(testCSSVarInCalc, dummyWindow{})
	if len(s.Groups) != 2 {
		t.FailNow()
	}
	if len(s.CustomVars) != 1 {
		t.FailNow()
	}
	if v, ok := s.CustomVars["--ed-menu-bar-height"]; !ok {
		t.FailNow()
	} else if len(v) != 1 {
		t.FailNow()
	} else if v[0] != "24px" {
		t.FailNow()
	}
	if len(s.Groups[1].Rules) != 1 {
		t.FailNow()
	}
	if s.Groups[1].Rules[0].Property != "height" {
		t.FailNow()
	}
	if len(s.Groups[1].Rules[0].Values) != 1 {
		t.FailNow()
	}
	if s.Groups[1].Rules[0].Values[0].Str != "calc" {
		t.FailNow()
	}
	if len(s.Groups[1].Rules[0].Values[0].Args) != 3 {
		t.FailNow()
	}
	expectedArgs := []string{"100%", "-", "24px"}
	for i := range expectedArgs {
		if s.Groups[1].Rules[0].Values[0].Args[i] != expectedArgs[i] {
			t.FailNow()
		}
	}
}
