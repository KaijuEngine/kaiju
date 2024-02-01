package spec_generator

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type pseudoData struct {
	name        string
	description string
	isFunction  bool
}

func (f pseudoData) StructName() string {
	titleCase := cases.Title(language.English)
	return strings.ReplaceAll(titleCase.String(strings.ReplaceAll(f.name, "-", " ")), " ", "")
}

var genPseudos = []pseudoData{
	{"active", "https://developer.mozilla.org/en-US/docs/Web/CSS/:active", false},
	{"any-link", "https://developer.mozilla.org/en-US/docs/Web/CSS/:any-link", false},
	{"autofill", "https://developer.mozilla.org/en-US/docs/Web/CSS/:autofill", false},
	{"blank", "https://developer.mozilla.org/en-US/docs/Web/CSS/:blank", false},
	{"checked", "https://developer.mozilla.org/en-US/docs/Web/CSS/:checked", false},
	{"current", "https://developer.mozilla.org/en-US/docs/Web/CSS/:current", false},
	{"default", "https://developer.mozilla.org/en-US/docs/Web/CSS/:default", false},
	{"defined", "https://developer.mozilla.org/en-US/docs/Web/CSS/:defined", false},
	{"dir", "https://developer.mozilla.org/en-US/docs/Web/CSS/:dir", true},
	{"disabled", "https://developer.mozilla.org/en-US/docs/Web/CSS/:disabled", false},
	{"empty", "https://developer.mozilla.org/en-US/docs/Web/CSS/:empty", false},
	{"enabled", "https://developer.mozilla.org/en-US/docs/Web/CSS/:enabled", false},
	{"first", "https://developer.mozilla.org/en-US/docs/Web/CSS/:first", false},
	{"first-child", "https://developer.mozilla.org/en-US/docs/Web/CSS/:first-child", false},
	{"first-of-type", "https://developer.mozilla.org/en-US/docs/Web/CSS/:first-of-type", false},
	{"fullscreen", "https://developer.mozilla.org/en-US/docs/Web/CSS/:fullscreen", false},
	{"future", "https://developer.mozilla.org/en-US/docs/Web/CSS/:future", false},
	{"focus", "https://developer.mozilla.org/en-US/docs/Web/CSS/:focus", false},
	{"focus-visible", "https://developer.mozilla.org/en-US/docs/Web/CSS/:focus-visible", false},
	{"focus-within", "https://developer.mozilla.org/en-US/docs/Web/CSS/:focus-within", false},
	{"has", "https://developer.mozilla.org/en-US/docs/Web/CSS/:has", true},
	{"host", "https://developer.mozilla.org/en-US/docs/Web/CSS/:host", true},
	{"host-context", "https://developer.mozilla.org/en-US/docs/Web/CSS/:host-context", true},
	{"hover", "https://developer.mozilla.org/en-US/docs/Web/CSS/:hover", false},
	{"indeterminate", "https://developer.mozilla.org/en-US/docs/Web/CSS/:indeterminate", false},
	{"in-range", "https://developer.mozilla.org/en-US/docs/Web/CSS/:in-range", false},
	{"invalid", "https://developer.mozilla.org/en-US/docs/Web/CSS/:invalid", false},
	{"is", "https://developer.mozilla.org/en-US/docs/Web/CSS/:is", true},
	{"lang", "https://developer.mozilla.org/en-US/docs/Web/CSS/:lang", true},
	{"last-child", "https://developer.mozilla.org/en-US/docs/Web/CSS/:last-child", false},
	{"last-of-type", "https://developer.mozilla.org/en-US/docs/Web/CSS/:last-of-type", false},
	{"left", "https://developer.mozilla.org/en-US/docs/Web/CSS/:left", false},
	{"link", "https://developer.mozilla.org/en-US/docs/Web/CSS/:link", false},
	{"local-link", "https://developer.mozilla.org/en-US/docs/Web/CSS/:local-link", false},
	{"modal", "https://developer.mozilla.org/en-US/docs/Web/CSS/:modal", false},
	{"not", "https://developer.mozilla.org/en-US/docs/Web/CSS/:not", true},
	{"nth-child", "https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-child", true},
	{"nth-col", "https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-col", true},
	{"nth-last-child", "https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-last-child", true},
	{"nth-last-col", "https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-last-col", true},
	{"nth-last-of-type", "https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-last-of-type", true},
	{"nth-of-type", "https://developer.mozilla.org/en-US/docs/Web/CSS/:nth-of-type", true},
	{"only-child", "https://developer.mozilla.org/en-US/docs/Web/CSS/:only-child", false},
	{"only-of-type", "https://developer.mozilla.org/en-US/docs/Web/CSS/:only-of-type", false},
	{"optional", "https://developer.mozilla.org/en-US/docs/Web/CSS/:optional", false},
	{"out-of-range", "https://developer.mozilla.org/en-US/docs/Web/CSS/:out-of-range", false},
	{"past", "https://developer.mozilla.org/en-US/docs/Web/CSS/:past", false},
	{"picture-in-picture", "https://developer.mozilla.org/en-US/docs/Web/CSS/:picture-in-picture", false},
	{"placeholder-shown", "https://developer.mozilla.org/en-US/docs/Web/CSS/:placeholder-shown", false},
	{"paused", "https://developer.mozilla.org/en-US/docs/Web/CSS/:paused", false},
	{"playing", "https://developer.mozilla.org/en-US/docs/Web/CSS/:playing", false},
	{"read-only", "https://developer.mozilla.org/en-US/docs/Web/CSS/:read-only", false},
	{"read-write", "https://developer.mozilla.org/en-US/docs/Web/CSS/:read-write", false},
	{"required", "https://developer.mozilla.org/en-US/docs/Web/CSS/:required", false},
	{"right", "https://developer.mozilla.org/en-US/docs/Web/CSS/:right", false},
	{"root", "https://developer.mozilla.org/en-US/docs/Web/CSS/:root", false},
	{"scope", "https://developer.mozilla.org/en-US/docs/Web/CSS/:scope", false},
	{"state", "https://developer.mozilla.org/en-US/docs/Web/CSS/:state", true},
	{"target", "https://developer.mozilla.org/en-US/docs/Web/CSS/:target", false},
	{"target-within", "https://developer.mozilla.org/en-US/docs/Web/CSS/:target-within", false},
	{"user-invalid", "https://developer.mozilla.org/en-US/docs/Web/CSS/:user-invalid", false},
	{"valid", "https://developer.mozilla.org/en-US/docs/Web/CSS/:valid", false},
	{"visited", "https://developer.mozilla.org/en-US/docs/Web/CSS/:visited", false},
	{"where", "https://developer.mozilla.org/en-US/docs/Web/CSS/:where", true},
}

func writePseudoFile() error {
	if err := writeBaseFile(pseudoFolder); err != nil {
		return err
	}
	pf, err := os.Create(pseudoFolder + "/css_pseudo.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package pseudos

import (
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

type Pseudo interface {
	Key() string
	IsFunction() bool
	Process(elm markup.DocumentElement, value rules.SelectorPart) ([]markup.DocumentElement, error)
}

var PseudoMap = map[string]Pseudo{
`)
	for _, p := range genPseudos {
		pf.WriteString(fmt.Sprintf(`	"%s": %s{},`, p.name, p.StructName()))
		pf.WriteString("\n")
	}
	pf.WriteString("}\n")
	return nil
}

func writePseudos() error {
	pf, err := os.Create(pseudoFolder + "/css_pseudo_types.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package pseudos
`)
	for _, f := range genPseudos {
		pf.WriteString(fmt.Sprintf(`
// %s
type %s struct{}

func (p %s) Key() string { return "%s" }
func (p %s) IsFunction() bool { return %v }
`, f.description, f.StructName(), f.StructName(), f.name, f.StructName(), f.isFunction))
	}
	for _, p := range genPseudos {
		fName := pseudoFolder + "/css_" + strings.ReplaceAll(p.name, "-", "_") + ".go"
		if _, err := os.Stat(fName); err != nil {
			if os.IsNotExist(err) {
				f, err := os.Create(fName)
				if err != nil {
					return err
				}
				defer f.Close()
				f.WriteString(fmt.Sprintf(`package pseudos

import (
	"errors"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (p %s) Process(elm markup.DocumentElement, value rules.SelectorPart) ([]markup.DocumentElement, error) {
	return []markup.DocumentElement{elm}, errors.New("not implemented")
}
`, p.StructName()))
			}
		}
	}
	return nil
}
