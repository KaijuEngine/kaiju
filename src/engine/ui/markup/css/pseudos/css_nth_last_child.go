/******************************************************************************/
/* css_nth_last_child.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p NthLastChild) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if start, skip, err := nth(value.Args, len(elm.Children)); err == nil {
		selected := make([]*document.Element, 0)
		for i := len(elm.Children) - 1 - start; i >= 0; i -= skip {
			selected = append(selected, elm.Children[i])
		}
		return selected, nil
	} else {
		return []*document.Element{}, err
	}
}
