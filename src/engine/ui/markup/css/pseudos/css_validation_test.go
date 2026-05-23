/******************************************************************************/
/* css_validation_test.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"testing"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func validationInput(t *testing.T, html string) *document.Element {
	t.Helper()
	root := document.NewHTML(html)
	elm := root.FindElementByTag("input")
	if elm == nil {
		t.Fatal("failed to create input test element")
	}
	return elm
}

func validationTextArea(t *testing.T, html string) *document.Element {
	t.Helper()
	root := document.NewHTML(html)
	elm := root.FindElementByTag("textarea")
	if elm == nil {
		t.Fatal("failed to create textarea test element")
	}
	return elm
}

func TestRequiredPseudoMatchesBooleanRequiredAttribute(t *testing.T) {
	elm := validationInput(t, `<input required>`)
	if elm.Attribute("required") != "" {
		t.Fatal("boolean required attribute should have an empty value")
	}
	if !elm.HasAttribute("required") {
		t.Fatal("required attribute presence should be detected")
	}
	got, err := (Required{}).Process(elm, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != elm {
		t.Fatal(":required should match inputs with a required attribute")
	}
}

func TestRequiredPseudoMatchesTextareaRequiredAttribute(t *testing.T) {
	elm := validationTextArea(t, `<textarea required></textarea>`)
	if elm.Attribute("required") != "" {
		t.Fatal("boolean required attribute should have an empty value")
	}
	if !elm.HasAttribute("required") {
		t.Fatal("required attribute presence should be detected")
	}
	got, err := (Required{}).Process(elm, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != elm {
		t.Fatal(":required should match textareas with a required attribute")
	}
}

func TestInvalidPseudoTargetsInputsWithPotentialConstraints(t *testing.T) {
	requiredInput := validationInput(t, `<input required>`)
	got, err := (Invalid{}).Process(requiredInput, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != requiredInput {
		t.Fatal(":invalid should be available for required inputs")
	}

	optionalInput := validationInput(t, `<input>`)
	got, err = (Invalid{}).Process(optionalInput, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != optionalInput {
		t.Fatal(":invalid rules should be attached to inputs and gated by runtime validity")
	}
}

func TestValidationPseudosTargetTextareas(t *testing.T) {
	requiredTextArea := validationTextArea(t, `<textarea required></textarea>`)
	got, err := (Invalid{}).Process(requiredTextArea, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != requiredTextArea {
		t.Fatal(":invalid should be available for required textareas")
	}

	got, err = (Valid{}).Process(requiredTextArea, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != requiredTextArea {
		t.Fatal(":valid rules should be attached to textareas and gated by runtime validity")
	}

	optionalTextArea := validationTextArea(t, `<textarea></textarea>`)
	got, err = (Optional{}).Process(optionalTextArea, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != optionalTextArea {
		t.Fatal(":optional should match textareas without required")
	}
}

func TestPlaceholderShownPseudoTargetsTextareaPlaceholders(t *testing.T) {
	elm := validationTextArea(t, `<textarea placeholder="Notes"></textarea>`)
	got, err := (PlaceholderShown{}).Process(elm, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != elm {
		t.Fatal(":placeholder-shown should match textarea placeholder candidates")
	}

	withoutPlaceholder := validationTextArea(t, `<textarea></textarea>`)
	got, err = (PlaceholderShown{}).Process(withoutPlaceholder, rules.SelectorPart{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatal(":placeholder-shown should skip textareas without a placeholder")
	}
}
