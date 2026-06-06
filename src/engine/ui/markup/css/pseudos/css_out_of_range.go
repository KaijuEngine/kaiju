/******************************************************************************/
/* css_out_of_range.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"errors"

	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p OutOfRange) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	return []*document.Element{elm}, errors.New("not implemented")
}
