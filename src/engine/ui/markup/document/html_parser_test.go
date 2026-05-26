/******************************************************************************/
/* html_parser_test.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package document

import "testing"

func TestClassifyHTMLInputType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want htmlInputType
	}{
		{"default empty", "", htmlInputTypeText},
		{"default unknown", "custom-widget", htmlInputTypeText},
		{"trim and case normalize", "  TeXt  ", htmlInputTypeText},
		{"search as text", "search", htmlInputTypeText},
		{"email", "email", htmlInputTypeEmail},
		{"url as text", "url", htmlInputTypeText},
		{"password", "password", htmlInputTypePassword},
		{"checkbox", "checkbox", htmlInputTypeCheckbox},
		{"slider alias", "slider", htmlInputTypeSlider},
		{"range alias", "range", htmlInputTypeSlider},
		{"number", "number", htmlInputTypeNumber},
		{"phone", "tel", htmlInputTypePhone},
		{"datetime", "datetime", htmlInputTypeDatetime},
		{"datetime-local", "datetime-local", htmlInputTypeDatetime},
		{"date", "date", htmlInputTypeDatetime},
		{"time", "time", htmlInputTypeDatetime},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyHTMLInputType(tt.in)
			if got != tt.want {
				t.Fatalf("classifyHTMLInputType(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseTextAreaInnerTextInitialValue(t *testing.T) {
	t.Parallel()

	textarea := mustFindTextArea(t, `<textarea>hello</textarea>`)
	if !textarea.IsTextArea() {
		t.Fatal("textarea tag should be identified as a textarea element")
	}
	if got := textarea.textAreaInitialValue(); got != "hello" {
		t.Fatalf("textAreaInitialValue() = %q, want %q", got, "hello")
	}
}

func TestParseTextAreaPlaceholderAndRequired(t *testing.T) {
	t.Parallel()

	textarea := mustFindTextArea(t, `<textarea placeholder="Say something" required></textarea>`)
	if got := textarea.Attribute("placeholder"); got != "Say something" {
		t.Fatalf("placeholder attribute = %q, want %q", got, "Say something")
	}
	if !textarea.HasAttribute("required") {
		t.Fatal("required attribute presence should be detected")
	}
}

func TestParseBooleanDisabledAttribute(t *testing.T) {
	t.Parallel()

	input := validationElementByTag(t, `<input disabled>`, "input")
	if input.Attribute("disabled") != "" {
		t.Fatal("boolean disabled attribute should have an empty value")
	}
	if !input.HasAttribute("disabled") {
		t.Fatal("disabled attribute presence should be detected")
	}
}

func TestSetElementDisabledUpdatesAttribute(t *testing.T) {
	t.Parallel()

	elm := validationElementByTag(t, `<button></button>`, "button")
	doc := Document{}

	doc.SetElementDisabled(elm, true)
	if !elm.HasAttribute("disabled") {
		t.Fatal("SetElementDisabled(true) should add disabled attribute")
	}

	doc.SetElementDisabled(elm, false)
	if elm.HasAttribute("disabled") {
		t.Fatal("SetElementDisabled(false) should remove disabled attribute")
	}
}

func TestParseTextAreaValueOverridesInnerText(t *testing.T) {
	t.Parallel()

	textarea := mustFindTextArea(t, `<textarea value="from value">from body</textarea>`)
	if got := textarea.textAreaInitialValue(); got != "from value" {
		t.Fatalf("textAreaInitialValue() = %q, want %q", got, "from value")
	}
}

func TestParseTextAreaTagLookup(t *testing.T) {
	t.Parallel()

	textarea := mustFindTextArea(t, `<div><textarea id="message" class="field note">hello</textarea></div>`)
	doc := Document{
		Elements:      []*Element{textarea},
		ids:           map[string]*Element{},
		classElements: map[string][]*Element{},
		tagElements:   map[string][]*Element{},
		groups:        map[string][]*Element{},
	}
	doc.reloadElementCaches()

	if got, ok := doc.GetElementById("message"); !ok || got != textarea {
		t.Fatal("textarea should be cached by id")
	}
	if got := doc.GetElementsByClass("field"); len(got) != 1 || got[0] != textarea {
		t.Fatal("textarea should be cached by class")
	}
	if got := doc.GetElementsByTagName("textarea"); len(got) != 1 || got[0] != textarea {
		t.Fatal("textarea should be cached by tag name")
	}
}

func TestSetElementStylePropertyWithoutApplyAddsProperty(t *testing.T) {
	t.Parallel()

	doc := Document{}
	elm := &Element{}

	doc.SetElementStylePropertyWithoutApply(elm, "color", "red")

	if got := elm.Attribute("style"); got != "color: red;" {
		t.Fatalf("style = %q, want %q", got, "color: red;")
	}
}

func TestSetElementStylePropertyWithoutApplyOverwritesProperty(t *testing.T) {
	t.Parallel()

	doc := Document{}
	elm := &Element{}
	elm.SetAttribute("style", "width: 10px; color: red; color: blue; height: 20px;")

	doc.SetElementStylePropertyWithoutApply(elm, "color", "green")

	want := "width: 10px; color: green; height: 20px;"
	if got := elm.Attribute("style"); got != want {
		t.Fatalf("style = %q, want %q", got, want)
	}
}

func TestSetElementStylePropertyWithoutApplyPreservesComplexValues(t *testing.T) {
	t.Parallel()

	doc := Document{}
	elm := &Element{}
	elm.SetAttribute("style", `background-image: url("data:image/svg+xml;utf8,<svg></svg>"); transform: translate(calc(100% - 4px), 0);`)

	doc.SetElementStylePropertyWithoutApply(elm, "opacity", "0.5;")

	want := `background-image: url("data:image/svg+xml;utf8,<svg></svg>"); transform: translate(calc(100% - 4px), 0); opacity: 0.5;`
	if got := elm.Attribute("style"); got != want {
		t.Fatalf("style = %q, want %q", got, want)
	}
}

func mustFindTextArea(t *testing.T, html string) *Element {
	t.Helper()
	return validationElementByTag(t, html, "textarea")
}

func validationElementByTag(t *testing.T, html string, tag string) *Element {
	t.Helper()
	root := NewHTML(html)
	elm := root.FindElementByTag(tag)
	if elm == nil {
		t.Fatalf("failed to parse %s element", tag)
	}
	return elm
}
