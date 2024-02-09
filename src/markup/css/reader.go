package css

import (
	"kaiju/engine"
	"kaiju/markup/css/properties"
	"kaiju/markup/css/pseudos"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

type CSSMap map[ui.UI][]rules.Rule

func (m CSSMap) add(elm ui.UI, rule []rules.Rule) {
	if _, ok := m[elm]; !ok {
		m[elm] = make([]rules.Rule, 0)
	}
	m[elm] = append(m[elm], rule...)
}

func applyToElement(inRules []rules.Rule, elm document.DocElement, host *engine.Host) []error {
	panel := elm.UIPanel
	hasHover := false
	for i := 0; i < len(inRules) && !hasHover; i++ {
		hasHover = inRules[i].Invocation == rules.RuleInvokeHover
	}
	proc := func(invokeType rules.RuleInvoke) []error {
		problems := make([]error, 0)
		for i := range inRules {
			rule := &inRules[i]
			if p, ok := properties.PropertyMap[rule.Property]; ok {
				if rule.Invocation == invokeType {
					if err := p.Process(panel, elm, rule.Values, host); err != nil {
						problems = append(problems, err)
					}
				}
			}
		}
		return problems
	}
	problems := proc(rules.RuleInvokeImmediate)
	if hasHover {
		elm.UI.AddEvent(ui.EventTypeEnter, func() {
			elm.UI.Layout().ClearFunctions()
			proc(rules.RuleInvokeImmediate)
			proc(rules.RuleInvokeHover)
		})
		elm.UI.AddEvent(ui.EventTypeExit, func() {
			elm.UI.Layout().ClearFunctions()
			proc(rules.RuleInvokeImmediate)
		})
	}
	return problems
}

func applyMappings(doc *document.Document, cssMap map[ui.UI][]rules.Rule, host *engine.Host) {
	for i := range doc.Elements {
		// TODO:  Make sure this is applying in order from parent to child
		// Since this array is intrinsically ordered, it should be fine
		if rules, ok := cssMap[doc.Elements[i].UI]; ok {
			applyToElement(rules, doc.Elements[i], host)
		}
	}
}

func applyDirect(part rules.SelectorPart, applyRules []rules.Rule, doc *document.Document, host *engine.Host, cssMap CSSMap) {
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

func applyIndirect(parts []rules.SelectorPart, applyRules []rules.Rule, doc *document.Document, host *engine.Host, cssMap CSSMap) {
	elms := make([]document.DocElement, 0)
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
	targets := make([]document.DocElement, 0)
	for _, part := range parts[1:] {
		for _, elm := range elms {
			if p, ok := pseudos.PseudoMap[part.Name]; ok {
				if selects, err := p.Process(elm, part); err == nil {
					targets = append(targets, selects...)
					applyRules = p.AlterRules(applyRules)
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

func cleanMapDuplicates(cssMap CSSMap) {
	for k, v := range cssMap {
		for i := range v {
			for j := i + 1; j < len(v); j++ {
				if v[i].Property == v[j].Property && v[i].Invocation == v[j].Invocation {
					v = append(v[:i], v[i+1:]...)
					i--
					break
				}
			}
		}
		cssMap[k] = v
	}
}

func Apply(s rules.StyleSheet, doc *document.Document, host *engine.Host) {
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
	cleanMapDuplicates(cssMap)
	applyMappings(doc, cssMap, host)
	for _, elm := range doc.Elements {
		if inlineStyle := elm.HTML.Attribute("style"); inlineStyle != "" {
			group := s.ParseInline(inlineStyle)
			applyToElement(group.Rules, elm, host)
		}
	}
}
