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

func mustFindTextArea(t *testing.T, html string) *Element {
	t.Helper()
	root := NewHTML(html)
	textarea := root.FindElementByTag("textarea")
	if textarea == nil {
		t.Fatal("failed to parse textarea element")
	}
	return textarea
}
