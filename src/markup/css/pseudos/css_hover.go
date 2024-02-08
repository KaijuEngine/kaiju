package pseudos

import (
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

func (p Hover) Process(elm document.DocElement, value rules.SelectorPart) ([]document.DocElement, error) {
	return []document.DocElement{elm}, nil
}

func (p Hover) AlterRules(inRules []rules.Rule) []rules.Rule {
	for i := range inRules {
		inRules[i].Invocation = rules.RuleInvokeHover
	}
	return inRules
}
