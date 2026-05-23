/******************************************************************************/
/* css_placeholder_shown.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p PlaceholderShown) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if isValidationControl(elm) && elm.HasAttribute("placeholder") {
		return []*document.Element{elm}, nil
	}
	return []*document.Element{}, nil
}
