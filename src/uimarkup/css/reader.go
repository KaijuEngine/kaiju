package css

import (
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/properties"
	"kaiju/uimarkup/css/pseudos"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

type CSSMap map[ui.UI][]rules.Rule

func (m CSSMap) add(elm ui.UI, rule []rules.Rule) {
	if _, ok := m[elm]; !ok {
		m[elm] = make([]rules.Rule, 0)
	}
	m[elm] = append(m[elm], rule...)
}

func applyToElement(rules []rules.Rule, elm markup.DocElement, host *engine.Host) []error {
	problems := make([]error, 0)
	panel := elm.UIPanel
	for _, rule := range rules {
		if p, ok := properties.PropertyMap[rule.Property]; ok {
			if err := p.Process(panel, elm, rule.Values, host); err != nil {
				problems = append(problems, err)
			}
		}
	}
	return problems
}

func applyMappings(doc *markup.Document, cssMap map[ui.UI][]rules.Rule, host *engine.Host) {
	for _, elm := range doc.Elements {
		// TODO:  Make sure this is applying in order from parent to child
		// Since this array is intrinsically ordered, it should be fine
		if rules, ok := cssMap[elm.UI]; ok {
			applyToElement(rules, elm, host)
		}
	}
}

func applyDirect(part rules.SelectorPart, applyRules []rules.Rule, doc *markup.Document, host *engine.Host, cssMap CSSMap) {
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

func applyIndirect(parts []rules.SelectorPart, applyRules []rules.Rule, doc *markup.Document, host *engine.Host, cssMap CSSMap) {
	elms := make([]markup.DocElement, 0)
	switch parts[0].SelectType {
	case rules.ReadingId:
		if elm, ok := doc.GetElementById(parts[0].Name); ok {
			elms = append(elms, elm)
		}
	case rules.ReadingClass:
		for _, elm := range doc.GetElementsByClass(parts[0].Name) {
			elms = append(elms, elm)
		}
	case rules.ReadingTag:
		for _, elm := range doc.GetElementsByTagName(parts[0].Name) {
			elms = append(elms, elm)
		}
	}
	targets := make([]markup.DocElement, 0)
	for _, part := range parts[1:] {
		for _, elm := range elms {
			if p, ok := pseudos.PseudoMap[part.Name]; ok {
				if selects, err := p.Process(elm, part); err == nil {
					targets = append(targets, selects...)
				}
			} else {
				tagged := doc.GetElementsByTagName(part.Name)
				for _, t := range tagged {
					if t.HTML.Parent == elm.HTML {
						targets = append(targets, t)
					}
				}
			}
		}
	}
	for _, target := range targets {
		cssMap.add(target.UI, applyRules)
	}
}

func Apply(s rules.StyleSheet, doc *markup.Document, host *engine.Host) {
	cssMap := CSSMap(make(map[ui.UI][]rules.Rule))
	for _, group := range s.Groups {
		for _, sel := range group.Selectors {
			if len(sel.Parts) == 1 {
				applyDirect(sel.Parts[0], group.Rules, doc, host, cssMap)
			} else if len(sel.Parts) > 1 {
				applyIndirect(sel.Parts, group.Rules, doc, host, cssMap)
			}
		}
	}
	applyMappings(doc, cssMap, host)
	for _, elm := range doc.Elements {
		if inlineStyle := elm.HTML.Attribute("style"); inlineStyle != "" {
			group := s.ParseInline(inlineStyle)
			applyToElement(group.Rules, elm, host)
		}
	}
}
