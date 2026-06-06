/******************************************************************************/
/* css_optional.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Optional) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if isValidationControl(elm) && !elm.HasAttribute("required") {
		return []*document.Element{elm}, nil
	}
	return []*document.Element{}, nil
}
