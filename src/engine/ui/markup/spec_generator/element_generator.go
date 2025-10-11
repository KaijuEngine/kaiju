/******************************************************************************/
/* element_generator.go                                                       */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
