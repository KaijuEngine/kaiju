package ui

import "testing"

func testInput(inputType InputType, text string, required bool) *Input {
	input := (*Input)(&UI{})
	input.elmData = &inputData{
		inputType: inputType,
		text:      text,
		required:  required,
	}
	return input
}

func TestInputTypeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputType InputType
		text      string
		required  bool
		want      bool
	}{
		{name: "optional empty email", inputType: InputTypeEmail, text: "", want: true},
		{name: "required empty email", inputType: InputTypeEmail, text: "", required: true, want: false},
		{name: "valid email", inputType: InputTypeEmail, text: "dev@example.com", want: true},
		{name: "email missing at", inputType: InputTypeEmail, text: "dev.example.com", want: false},
		{name: "email with space", inputType: InputTypeEmail, text: "dev @example.com", want: false},
		{name: "valid number", inputType: InputTypeNumber, text: "-12.5", want: true},
		{name: "valid exponent number", inputType: InputTypeNumber, text: "1e3", want: true},
		{name: "invalid number", inputType: InputTypeNumber, text: "12abc", want: false},
		{name: "valid tel", inputType: InputTypePhone, text: "+1 (555) 010-1234", want: true},
		{name: "tel needs a digit", inputType: InputTypePhone, text: "+ --", want: false},
		{name: "tel rejects letters", inputType: InputTypePhone, text: "555-CALL", want: false},
		{name: "password has no format validation", inputType: InputTypePassword, text: "secret phrase", want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := testInput(tt.inputType, tt.text, tt.required)
			if got := input.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInputTypeSanitizesInsertedText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputType InputType
		text      string
		want      string
	}{
		{name: "email removes whitespace", inputType: InputTypeEmail, text: "dev @example.com\n", want: "dev@example.com"},
		{name: "number keeps numeric characters", inputType: InputTypeNumber, text: "a-12.5e3b", want: "-12.5e3"},
		{name: "tel keeps phone characters", inputType: InputTypePhone, text: "call +1 (555) 010-1234", want: " +1 (555) 010-1234"},
		{name: "password preserves text", inputType: InputTypePassword, text: "secret phrase!", want: "secret phrase!"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := testInput(tt.inputType, "", false)
			if got := input.sanitizeText(tt.text); got != tt.want {
				t.Fatalf("sanitizeText(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestPasswordInputMasksDisplayText(t *testing.T) {
	t.Parallel()

	input := testInput(InputTypePassword, "", false)
	if got := input.displayText("secret"); got != "******" {
		t.Fatalf("displayText() = %q, want %q", got, "******")
	}

	text := testInput(InputTypeText, "", false)
	if got := text.displayText("secret"); got != "secret" {
		t.Fatalf("text input displayText() = %q, want %q", got, "secret")
	}
}
