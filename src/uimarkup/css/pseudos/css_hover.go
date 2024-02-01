package pseudos

import (
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (p Hover) Process(elm markup.DocElement, value rules.SelectorPart) ([]markup.DocElement, error) {
	return []markup.DocElement{elm}, nil
}

func (p Hover) AlterRules(inRules []rules.Rule) []rules.Rule {
	for i := range inRules {
		inRules[i].Invocation = rules.RuleInvokeHover
	}
	return inRules
}
