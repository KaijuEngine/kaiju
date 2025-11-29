/******************************************************************************/
/* reader.go                                                                  */
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

package css

import (
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/pseudos"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/platform/windowing"
	"slices"
)

type CSSMap map[*ui.UI][]rules.Rule

func (m CSSMap) add(elm *ui.UI, rule []rules.Rule) {
	if _, ok := m[elm]; !ok {
		m[elm] = make([]rules.Rule, 0)
	}
	m[elm] = append(m[elm], rule...)
}

func applyToElement(inRules []rules.Rule, elm *document.Element) {
	for i := range inRules {
		elm.Stylizer.AddRule(inRules[i].Clone())
	}
	elm.UI.Layout().Stylizer = &elm.Stylizer
	elm.UI.SetDirty(ui.DirtyTypeGenerated)
}

func applyMappings(doc *document.Document, cssMap map[*ui.UI][]rules.Rule) {
	for i := range doc.Elements {
		// TODO:  Make sure this is applying in order from parent to child
		// Since this array is intrinsically ordered, it should be fine
		if rules, ok := cssMap[doc.Elements[i].UI]; ok {
			applyToElement(rules, doc.Elements[i])
		}
	}
}

func applyDirect(part rules.SelectorPart, applyRules []rules.Rule, doc *document.Document, cssMap CSSMap) {
	switch part.SelectType {
	case rules.ReadingId:
		if elm, ok := doc.GetElementById(part.Name); ok {
			cssMap.add(elm.UI, applyRules)
		}
	case rules.ReadingClass:
		for _, elm := range doc.GetElementsByClass(part.Name) {
			cssMap.add(elm.UI, applyRules)
		}
	case rules.ReadingTag:
		for _, elm := range doc.GetElementsByTagName(part.Name) {
			cssMap.add(elm.UI, applyRules)
		}
	}
}

func applyIndirect(parts []rules.SelectorPart, applyRules []rules.Rule, doc *document.Document, cssMap CSSMap) {
	elms := make([]*document.Element, 0)
	switch parts[0].SelectType {
	case rules.ReadingId:
		if elm, ok := doc.GetElementById(parts[0].Name); ok {
			elms = append(elms, elm)
		}
	case rules.ReadingClass:
		elms = append(elms, doc.GetElementsByClass(parts[0].Name)...)
	case rules.ReadingTag:
		elms = append(elms, doc.GetElementsByTagName(parts[0].Name)...)
	}
	targets := make([]*document.Element, 0)
	lastTargets := []*document.Element{}
	for _, elm := range elms {
		lastTargets = append(lastTargets, elm)
		for _, part := range parts[1:] {
			if p, ok := pseudos.PseudoMap[part.Name]; ok {
				for i := range lastTargets {
					if selects, err := p.Process(lastTargets[i], part); err == nil {
						targets = klib.AppendUnique(targets, selects...)
						applyRules = p.AlterRules(applyRules)
					}
				}
			} else {
				if part.Name == "wide" {
					println("test")
				}
				switch part.SelectType {
				case rules.ReadingClass:
					if elm.HasClass(part.Name) {
						targets = append(targets, elm)
					}
				case rules.ReadingTag:
					tagged := doc.GetElementsByTagName(part.Name)
					lastTargets = lastTargets[:0]
					for _, t := range tagged {
						if t.Parent.Value() == elm {
							targets = append(targets, t)
							lastTargets = append(lastTargets, t)
						}
					}
				}
			}
		}
	}
	for _, target := range targets {
		cssMap.add(target.UI, applyRules)
	}
}

func cleanMapDuplicates(cssMap CSSMap) {
	for k, v := range cssMap {
		for i := 0; i < len(v); i++ {
			for j := i + 1; j < len(v); j++ {
				if v[i].Property == v[j].Property && v[i].Invocation == v[j].Invocation {
					v = slices.Delete(v, i, i+1)
					i--
					break
				}
			}
		}
		cssMap[k] = v
	}
}

type Stylizer struct {
	Window *windowing.Window
}

func (z Stylizer) ApplyStyles(s rules.StyleSheet, doc *document.Document) {
	for i := range doc.Elements {
		e := doc.Elements[i]
		e.Stylizer.ClearRules()
		for j := range e.UIEventIds {
			for k := range e.UIEventIds[j] {
				e.UI.RemoveEvent(j, e.UIEventIds[j][k])
			}
		}
	}
	cssMap := CSSMap(make(map[*ui.UI][]rules.Rule))
	for _, group := range s.Groups {
		for _, sel := range group.Selectors {
			if len(sel.Parts) == 1 {
				applyDirect(sel.Parts[0], group.Rules, doc, cssMap)
			} else if len(sel.Parts) > 1 {
				applyIndirect(sel.Parts, group.Rules, doc, cssMap)
			}
		}
	}
	cleanMapDuplicates(cssMap)
	applyMappings(doc, cssMap)
	for _, elm := range doc.Elements {
		if inlineStyle := elm.Attribute("style"); inlineStyle != "" {
			group := s.ParseInline(inlineStyle, z.Window)
			applyToElement(group.Rules, elm)
		}
	}
}
