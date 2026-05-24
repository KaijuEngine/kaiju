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

type selectorStep struct {
	parts      []rules.SelectorPart
	combinator rules.RuleState
}

func isCombinator(part rules.SelectorPart) bool {
	switch part.SelectType {
	case rules.ReadingDescendant, rules.ReadingChild, rules.ReadingSibling, rules.ReadingAdjacent:
		return true
	default:
		return false
	}
}

func selectorSteps(parts []rules.SelectorPart) []selectorStep {
	steps := make([]selectorStep, 0, len(parts))
	current := selectorStep{parts: make([]rules.SelectorPart, 0), combinator: rules.ReadingTag}
	for i := range parts {
		part := parts[i]
		if isCombinator(part) {
			if len(current.parts) > 0 {
				steps = append(steps, current)
				current = selectorStep{parts: make([]rules.SelectorPart, 0), combinator: part.SelectType}
			} else if len(steps) > 0 {
				current.combinator = part.SelectType
			}
			continue
		}
		current.parts = append(current.parts, part)
	}
	if len(current.parts) > 0 {
		steps = append(steps, current)
	}
	return steps
}

func selectorMatches(elm *document.Element, parts []rules.SelectorPart, applyRules []rules.Rule) (bool, []rules.Rule) {
	steps := selectorSteps(parts)
	if len(steps) == 0 {
		return false, nil
	}
	selectorRules := rules.CloneRules(applyRules)
	if selectorStepMatches(elm, steps, len(steps)-1, &selectorRules) {
		return true, selectorRules
	}
	return false, nil
}

func selectorStepMatches(elm *document.Element, steps []selectorStep, idx int, selectorRules *[]rules.Rule) bool {
	if elm == nil || !selectorPartListMatches(elm, steps[idx].parts, selectorRules) {
		return false
	}
	if idx == 0 {
		return true
	}
	switch steps[idx].combinator {
	case rules.ReadingChild:
		return selectorStepMatches(elm.Parent.Value(), steps, idx-1, selectorRules)
	case rules.ReadingDescendant:
		for parent := elm.Parent.Value(); parent != nil; parent = parent.Parent.Value() {
			if selectorStepMatches(parent, steps, idx-1, selectorRules) {
				return true
			}
		}
	}
	return false
}

func selectorPartListMatches(elm *document.Element, parts []rules.SelectorPart, selectorRules *[]rules.Rule) bool {
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		switch part.SelectType {
		case rules.ReadingId:
			if elm.Attribute("id") != part.Name {
				return false
			}
		case rules.ReadingClass:
			if !elm.HasClass(part.Name) {
				return false
			}
		case rules.ReadingTag:
			if elm.IsText() || elm.Data != part.Name {
				return false
			}
		case rules.ReadingCondition:
			want := ""
			hasAssignment := i+1 < len(parts) && parts[i+1].SelectType == rules.ReadingConditionAssignment
			if hasAssignment {
				want = parts[i+1].Name
				i++
			}
			if !selectorAttributeMatches(elm, part.Name, want, hasAssignment) {
				return false
			}
		case rules.ReadingConditionAssignment:
			return false
		case rules.ReadingPseudo, rules.ReadingPseudoFunction:
			p, ok := pseudos.PseudoMap[part.Name]
			if !ok {
				return false
			}
			selects, err := p.Process(elm, part)
			if err != nil || !slices.Contains(selects, elm) {
				return false
			}
			*selectorRules = p.AlterRules(*selectorRules)
		}
	}
	return true
}

func selectorAttributeMatches(elm *document.Element, key, value string, hasAssignment bool) bool {
	if !hasAssignment {
		return elm.HasAttribute(key)
	}
	return elm.Attribute(key) == value
}

func applyIndirect(parts []rules.SelectorPart, applyRules []rules.Rule, doc *document.Document, cssMap CSSMap) {
	for _, elm := range doc.Elements {
		if ok, selectorRules := selectorMatches(elm, parts, applyRules); ok {
			cssMap.add(elm.UI, selectorRules)
		}
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
			if len(sel.Parts) == 1 && (sel.Parts[0].SelectType == rules.ReadingId ||
				sel.Parts[0].SelectType == rules.ReadingClass ||
				sel.Parts[0].SelectType == rules.ReadingTag) {
				applyDirect(sel.Parts[0], group.Rules, doc, cssMap)
			} else if len(sel.Parts) > 1 {
				applyIndirect(sel.Parts, group.Rules, doc, cssMap)
			} else if len(sel.Parts) == 1 {
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
