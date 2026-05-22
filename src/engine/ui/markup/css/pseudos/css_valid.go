/******************************************************************************/
/* css_valid.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Valid) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if elm.Data == "input" {
		return []*document.Element{elm}, nil
	}
	return []*document.Element{}, nil
}

func (p Valid) AlterRules(inRules []rules.Rule) []rules.Rule {
	for i := range inRules {
		inRules[i].Invocation = inRules[i].Invocation.With(rules.RuleInvokeValid)
	}
	return inRules
}
