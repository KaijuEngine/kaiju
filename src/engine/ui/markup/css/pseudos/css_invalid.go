/******************************************************************************/
/* css_invalid.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Invalid) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if isValidationControl(elm) {
		return []*document.Element{elm}, nil
	}
	return []*document.Element{}, nil
}

func (p Invalid) AlterRules(inRules []rules.Rule) []rules.Rule {
	for i := range inRules {
		inRules[i].Invocation = inRules[i].Invocation.With(rules.RuleInvokeInvalid)
	}
	return inRules
}
