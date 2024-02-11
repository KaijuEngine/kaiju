/*****************************************************************************/
/* function_generator.go                                                     */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md)    */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining     */
/* a copy of this software and associated documentation files (the           */
/* "Software"), to deal in the Software without restriction, including       */
/* without limitation the rights to use, copy, modify, merge, publish,       */
/* distribute, sublicense, and/or sell copies of the Software, and to        */
/* permit persons to whom the Software is furnished to do so, subject to     */
/* the following conditions:                                                 */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,           */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF        */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY      */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,      */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE         */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                    */
/*****************************************************************************/

package spec_generator

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type functionData struct {
	name        string
	description string
}

func (f functionData) StructName() string {
	titleCase := cases.Title(language.English)
	return strings.ReplaceAll(titleCase.String(strings.ReplaceAll(f.name, "-", " ")), " ", "")
}

var genFuncs = []functionData{
	{"attr", "Returns the value of an attribute of the selected element"},
	{"calc", "Allows you to perform calculations to determine CSS property values"},
	{"conic-gradient", "Creates a conic gradient"},
	{"counter", "Returns the current value of the named counter"},
	{"cubic-bezier", "Defines a Cubic Bezier curve"},
	{"hsl", "Defines colors using the Hue-Saturation-Lightness model (HSL)"},
	{"hsla", "Defines colors using the Hue-Saturation-Lightness-Alpha model (HSLA)"},
	{"linear-gradient", "Creates a linear gradient"},
	{"max", "Uses the largest value, from a comma-separated list of values, as the property value"},
	{"min", "Uses the smallest value, from a comma-separated list of values, as the property value"},
	{"radial-gradient", "Creates a radial gradient"},
	{"repeating-conic-gradient", "Repeats a conic gradient"},
	{"repeating-linear-gradient", "Repeats a linear gradient"},
	{"repeating-radial-gradient", "Repeats a radial gradient"},
	{"rgb", "Defines colors using the Red-Green-Blue model (RGB)"},
	{"rgba", "Defines colors using the Red-Green-Blue-Alpha model (RGBA)"},
	{"var", "Inserts the value of a custom property"},
}

func writeFunctionFile() error {
	if err := writeBaseFile(funcFolder); err != nil {
		return err
	}
	pf, err := os.Create(funcFolder + "/css_function.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package functions

import (
	"kaiju/ui"
	"kaiju/markup/css/rules"
	"kaiju/markup/markup"
)

type Function interface {
	Key() string
	Process(panel *ui.Panel, elm document.DocumentElement, value rules.PropertyValue) (string, error)
}

var FunctionMap = map[string]Function{
`)
	for _, p := range genFuncs {
		pf.WriteString(fmt.Sprintf(`	"%s": %s{},`, p.name, p.StructName()))
		pf.WriteString("\n")
	}
	pf.WriteString("}\n")
	return nil
}

func writeFunctions() error {
	pf, err := os.Create(funcFolder + "/css_function_types.go")
	if err != nil {
		return err
	}
	defer pf.Close()
	pf.WriteString(`package functions
`)
	for _, f := range genFuncs {
		pf.WriteString(fmt.Sprintf(`
// %s
type %s struct{}

func (f %s) Key() string { return "%s" }
`, f.description, f.StructName(), f.StructName(), f.name))
	}
	for _, p := range genFuncs {
		fName := funcFolder + "/css_" + strings.ReplaceAll(p.name, "-", "_") + ".go"
		if _, err := os.Stat(fName); err != nil {
			if os.IsNotExist(err) {
				f, err := os.Create(fName)
				if err != nil {
					return err
				}
				defer f.Close()
				f.WriteString(fmt.Sprintf(`package functions

import (
	"errors"
	"kaiju/ui"
	"kaiju/markup/css/rules"
	"kaiju/markup/markup"
)

func (f %s) Process(panel *ui.Panel, elm document.DocumentElement, value rules.PropertyValue) (string, error) {
	return "", errors.New("not implemented")
}
`, p.StructName()))
			}
		}
	}
	return nil
}
