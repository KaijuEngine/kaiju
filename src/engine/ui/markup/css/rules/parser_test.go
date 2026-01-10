/******************************************************************************/
/* parser_test.go                                                             */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package rules

import "testing"

const testCSSNarrowTag = `.entry span { display: none; }`
const testCSSNarrowClass = `.entry .wide { display: none; }`
const testCSSCommaId = `#id1, #id2, #id3 { display: none; }`
const testCSSInputText = `input[type="text"] { display: none; }`
const testCSSVarDeclare = `:root { --ed-menu-bar-height: 24px; }
.test { height: var(--ed-menu-bar-height); }`
const testCSSVarInCalc = `:root { --ed-menu-bar-height: 24px; }
.test { height: calc(100% - var(--ed-menu-bar-height)); }`

type dummyWindow struct{}

func (dummyWindow) DotsPerMillimeter() float64 { return 1 }

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
	if len(sel.Parts) != 2 {
		t.FailNow()
	}
	if sel.Parts[0].Name != "entry" {
		t.FailNow()
	}
	if sel.Parts[0].SelectType != ReadingClass {
		t.FailNow()
	}
	if sel.Parts[1].Name != "span" {
		t.FailNow()
	}
	if sel.Parts[1].SelectType != ReadingTag {
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
	if len(sel.Parts) != 2 {
		t.FailNow()
	}
	if sel.Parts[0].Name != "entry" {
		t.FailNow()
	}
	if sel.Parts[0].SelectType != ReadingClass {
		t.FailNow()
	}
	if sel.Parts[1].Name != "wide" {
		t.FailNow()
	}
	if sel.Parts[1].SelectType != ReadingClass {
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
