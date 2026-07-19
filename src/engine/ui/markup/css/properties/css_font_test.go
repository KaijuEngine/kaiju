/******************************************************************************/
/* css_font_test.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"testing"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/rendering"
)

func TestInheritedTextPropertiesSkipElementsWithoutUI(t *testing.T) {
	// Select options are represented in the document tree but intentionally do
	// not own a UI object. Inherited text properties must safely cross them.
	optionText := &document.Element{}
	option := &document.Element{Children: []*document.Element{optionText}}
	root := &document.Element{Children: []*document.Element{option}}

	setChildrenFontFace(root, rendering.FontRegular)
	setChildrenFontSize(root, "12px", nil)
	setChildrenFontStyle(root, "normal")
	setChildrenFontWeight(root, "normal")
	setChildrenLineHeight(root, "normal", nil)
	setChildTextWordWrap(root, true)

	if err := (FontSize{}).Process(nil, option, []rules.PropertyValue{{Str: "12px"}}, nil); err != nil {
		t.Fatal(err)
	}
	if err := (LineHeight{}).Process(nil, option, []rules.PropertyValue{{Str: "normal"}}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestParseFontShorthandFull(t *testing.T) {
	values := []rules.PropertyValue{
		{Str: "italic"},
		{Str: "small-caps"},
		{Str: "bold"},
		{Str: "24px", Num: 24},
		{Str: "/"},
		{Str: "120%", Num: 1.2},
		{Str: "'OpenSans-Regular'"},
	}
	got, err := parseFontShorthand(values)
	if err != nil {
		t.Fatalf("parseFontShorthand returned error: %v", err)
	}
	if got.style != "italic" {
		t.Fatalf("style = %q, want italic", got.style)
	}
	if got.weight != "bold" {
		t.Fatalf("weight = %q, want bold", got.weight)
	}
	if got.size.Str != "24px" {
		t.Fatalf("size = %q, want 24px", got.size.Str)
	}
	if !got.hasLine || got.line.Str != "120%" {
		t.Fatalf("line = (%v, %q), want present 120%%", got.hasLine, got.line.Str)
	}
	if len(got.family) != 1 || got.family[0].Str != "'OpenSans-Regular'" {
		t.Fatalf("family = %#v, want OpenSans-Regular", got.family)
	}
}

func TestParseFontShorthandDefaultsStyleAndWeight(t *testing.T) {
	values := []rules.PropertyValue{
		{Str: "18px", Num: 18},
		{Str: "'OpenSans-Regular'"},
	}
	got, err := parseFontShorthand(values)
	if err != nil {
		t.Fatalf("parseFontShorthand returned error: %v", err)
	}
	if got.style != "normal" {
		t.Fatalf("style = %q, want normal", got.style)
	}
	if got.weight != "normal" {
		t.Fatalf("weight = %q, want normal", got.weight)
	}
}

func TestParseFontShorthandRequiresSizeAndFamily(t *testing.T) {
	tests := [][]rules.PropertyValue{
		{{Str: "bold"}, {Str: "'OpenSans-Regular'"}},
		{{Str: "bold"}, {Str: "18px", Num: 18}},
	}
	for _, values := range tests {
		if _, err := parseFontShorthand(values); err == nil {
			t.Fatalf("parseFontShorthand(%#v) expected error", values)
		}
	}
}
