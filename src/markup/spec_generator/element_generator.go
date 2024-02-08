package spec_generator

import (
	"fmt"
	"os"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type elementData struct {
	name string
}

func (e elementData) StructName() string {
	titleCase := cases.Title(language.English)
	return titleCase.String(e.name)
}

var genElms = []elementData{
	{"a"},
	{"abbr"},
	{"address"},
	{"area"},
	{"article"},
	{"aside"},
	{"audio"},
	{"b"},
	{"base"},
	{"bdi"},
	{"bdo"},
	{"blockquote"},
	{"body"},
	{"br"},
	{"button"},
	{"canvas"},
	{"caption"},
	{"cite"},
	{"code"},
	{"col"},
	{"colgroup"},
	{"data"},
	{"datalist"},
	{"dd"},
	{"del"},
	{"details"},
	{"dfn"},
	{"dialog"},
	{"div"},
	{"dl"},
	{"dt"},
	{"em"},
	{"embed"},
	{"fieldset"},
	{"figcaption"},
	{"figure"},
	{"footer"},
	{"form"},
	{"h1"},
	{"h2"},
	{"h3"},
	{"h4"},
	{"h5"},
	{"h6"},
	{"head"},
	{"header"},
	{"hgroup"},
	{"hr"},
	{"html"},
	{"i"},
	{"iframe"},
	{"img"},
	{"input"},
	{"ins"},
	{"kbd"},
	{"label"},
	{"legend"},
	{"li"},
	{"link"},
	{"main"},
	{"map"},
	{"mark"},
	{"menu"},
	{"meta"},
	{"meter"},
	{"nav"},
	{"noscript"},
	{"object"},
	{"ol"},
	{"optgroup"},
	{"option"},
	{"output"},
	{"p"},
	{"picture"},
	{"pre"},
	{"progress"},
	{"q"},
	{"rp"},
	{"rt"},
	{"ruby"},
	{"s"},
	{"samp"},
	{"script"},
	{"search"},
	{"section"},
	{"select"},
	{"slot"},
	{"small"},
	{"source"},
	{"span"},
	{"strong"},
	{"style"},
	{"sub"},
	{"summary"},
	{"sup"},
	{"table"},
	{"tbody"},
	{"td"},
	{"template"},
	{"textarea"},
	{"tfoot"},
	{"th"},
	{"thead"},
	{"time"},
	{"title"},
	{"tr"},
	{"track"},
	{"u"},
	{"ul"},
	{"var"},
	{"video"},
	{"wbr"},
}

func writeElementsFile() error {
	if err := writeBaseFile(elmFolder); err != nil {
		return err
	}
	pf, err := os.Create(elmFolder + "/html_element.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package elements

type Element interface {
	Key() string
}

var ElementMap = map[string]Element{
`)
	for _, p := range genElms {
		pf.WriteString(fmt.Sprintf(`	"%s": %s{},`, p.name, p.StructName()))
		pf.WriteString("\n")
	}
	pf.WriteString("}\n")
	return nil
}

func writeElements() error {
	pf, err := os.Create(elmFolder + "/html_element_types.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package elements
`)
	for _, e := range genElms {
		pf.WriteString(fmt.Sprintf(`
type %s struct{}

func (p %s) Key() string { return "%s" }
`, e.StructName(), e.StructName(), e.name))
	}
	return nil
}
