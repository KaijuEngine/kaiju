/******************************************************************************/
/* reader.go                                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package css

import (
	"slices"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/pseudos"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/windowing"
)

type CSSMap map[*ui.UI][]rules.Rule

var cssShorthandLonghands = map[string]map[string]struct{}{
	"margin": {
		"margin-top":    {},
		"margin-right":  {},
		"margin-bottom": {},
		"margin-left":   {},
	},
	"padding": {
		"padding-top":    {},
		"padding-right":  {},
		"padding-bottom": {},
		"padding-left":   {},
	},
	"border": {
		"border-top":          {},
		"border-right":        {},
		"border-bottom":       {},
		"border-left":         {},
		"border-top-width":    {},
		"border-right-width":  {},
		"border-bottom-width": {},
		"border-left-width":   {},
		"border-top-style":    {},
		"border-right-style":  {},
		"border-bottom-style": {},
		"border-left-style":   {},
		"border-top-color":    {},
		"border-right-color":  {},
		"border-bottom-color": {},
		"border-left-color":   {},
		"border-width":        {},
		"border-style":        {},
		"border-color":        {},
	},
	"border-width": {
		"border-top-width":    {},
		"border-right-width":  {},
		"border-bottom-width": {},
		"border-left-width":   {},
	},
	"border-style": {
		"border-top-style":    {},
		"border-right-style":  {},
		"border-bottom-style": {},
		"border-left-style":   {},
	},
	"border-color": {
		"border-top-color":    {},
		"border-right-color":  {},
		"border-bottom-color": {},
		"border-left-color":   {},
	},
	"border-radius": {
		"border-top-left-radius":     {},
		"border-top-right-radius":    {},
		"border-bottom-right-radius": {},
		"border-bottom-left-radius":  {},
	},
}

func cssPropertyOverrides(later, earlier string) bool {
	if later == earlier {
		return true
	}
	if longhands, ok := cssShorthandLonghands[later]; ok {
		_, ok = longhands[earlier]
		return ok
	}
	return false
}

func (m CSSMap) add(elm *ui.UI, inRules []rules.Rule) {
	addRules := rules.CloneRules(inRules)
	if c, ok := m[elm]; !ok {
		m[elm] = addRules
	} else {
		for i := len(c) - 1; i >= 0; i-- {
			for j := range addRules {
				if c[i].Invocation == addRules[j].Invocation &&
					cssPropertyOverrides(addRules[j].Property, c[i].Property) {
					c = slices.Delete(c, i, i+1)
					break
				}
			}
		}
		c = append(c, addRules...)
		m[elm] = c
	}
}

func applyToElement(inRules []rules.Rule, elm *document.Element) {
	for i := range inRules {
		elm.Stylizer.AddRule(inRules[i].Clone())
	}
	elm.UI.Layout().Stylizer = &elm.Stylizer
	elm.UI.SetDirty(ui.DirtyTypeGenerated)
}

func applyMappings(doc *document.Document, cssMap map[*ui.UI][]rules.Rule) {
	for _, e := range doc.Elements {
		// TODO:  Make sure this is applying in order from parent to child
		// Since this array is intrinsically ordered, it should be fine
		if rules, ok := cssMap[e.UI]; ok {
			applyToElement(rules, e)
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
	selectorRules := rules.CloneRules(applyRules)
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
	partsLoop:
		for _, part := range parts[1:] {
			switch part.SelectType {
			case rules.ReadingCondition:
				if len(parts) > 2 && elm.Attribute(parts[1].Name) == parts[2].Name {
					targets = klib.AppendUnique(targets, elm)
				} else {
					break partsLoop
				}
			case rules.ReadingClass:
				if elm.HasClass(part.Name) {
					targets = klib.AppendUnique(targets, elm)
				}
			case rules.ReadingTag:
				tagged := doc.GetElementsByTagName(part.Name)
				lastTargets = lastTargets[:0]
				for _, t := range tagged {
					if t.Parent.Value() == elm {
						targets = klib.AppendUnique(targets, t)
						lastTargets = append(lastTargets, t)
					}
				}
			case rules.ReadingPseudo, rules.ReadingPseudoFunction:
				if p, ok := pseudos.PseudoMap[part.Name]; ok {
					for i := range lastTargets {
						if selects, err := p.Process(lastTargets[i], part); err == nil {
							targets = klib.AppendUnique(targets, selects...)
							selectorRules = p.AlterRules(selectorRules)
						}
					}
				}
			}
		}
	}
	for _, target := range targets {
		cssMap.add(target.UI, selectorRules)
	}
}

func cleanMapDuplicates(cssMap CSSMap) {
	for k, v := range cssMap {
		for i := 0; i < len(v); i++ {
			if len(v[i].Values) == 1 && v[i].Values[0].Str == "revert" {
				v = slices.Delete(v, i, i+1)
				i--
				continue
			}
			for j := i + 1; j < len(v); j++ {
				if v[i].Invocation == v[j].Invocation &&
					cssPropertyOverrides(v[j].Property, v[i].Property) {
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
			e.UIEventIds[j] = e.UIEventIds[j][:0]
		}
	}
	cssMap := CSSMap(make(map[*ui.UI][]rules.Rule))
	for _, group := range s.Groups {
		if group.MediaQuery.IsValid() {
			switch group.MediaQuery.Key {
			case "screen":
			case "max-width":
				v := helpers.NumFromLength(group.MediaQuery.Value, z.Window)
				if int(v) <= z.Window.Width() {
					continue
				}
			default:
				continue
			}
		}
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
