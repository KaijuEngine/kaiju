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
