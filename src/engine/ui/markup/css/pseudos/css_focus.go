/******************************************************************************/
/* css_focus.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Focus) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	return []*document.Element{elm}, nil
}

func (p Focus) AlterRules(inRules []rules.Rule) []rules.Rule {
	for i := range inRules {
		inRules[i].Invocation = inRules[i].Invocation.With(rules.RuleInvokeFocus)
	}
	return inRules
}
