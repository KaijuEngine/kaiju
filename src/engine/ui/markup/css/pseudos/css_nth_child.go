/******************************************************************************/
/* css_nth_child.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package pseudos

import (
	"errors"
	"fmt"

	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func nth(args []string, count int) (int, int, error) {
	if len(args) == 0 {
		return 0, 0, errors.New("no arguments supplied")
	} else if count == 0 {
		return 0, 0, errors.New("no children")
	} else {
		start := 0
		skip := 0
		prevArg := args[0]
		defer func() { args[0] = prevArg }()
		var err error
		switch args[0] {
		case "odd":
			start = 1
			fallthrough
		case "even":
			args[0] = "2"
		}
		helpers.ChangeNToChildCount(args, count)
		if skip, err = helpers.ArithmeticString(args); err != nil {
			return 0, 0, err
		} else if skip <= 0 {
			return 0, 0, fmt.Errorf("invalid skip value: %d", skip)
		}
		return start, skip, nil
	}
}

func (p NthChild) Process(elm *document.Element, value rules.SelectorPart) ([]*document.Element, error) {
	if start, skip, err := nth(value.Args, len(elm.Children)); err == nil {
		selected := make([]*document.Element, 0)
		for i := start; i < len(elm.Children); i += skip {
			selected = append(selected, elm.Children[i])
		}
		return selected, nil
	} else {
		return []*document.Element{}, err
	}
}
