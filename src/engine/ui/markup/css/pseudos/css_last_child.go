/******************************************************************************/
/* css_last_child.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"errors"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p LastChild) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if len(elm.Children) == 0 {
		return []*document.Element{}, errors.New("no children")
	} else {
		idx := len(elm.Children) - 1
		return []*document.Element{elm.Children[idx]}, nil
	}
}
