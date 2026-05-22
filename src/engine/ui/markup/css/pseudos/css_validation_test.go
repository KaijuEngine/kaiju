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

func TestInvalidPseudoOnlyTargetsRequiredInputs(t *testing.T) {
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
	if len(got) != 0 {
		t.Fatal(":invalid should not target optional inputs without constraints")
	}
}
