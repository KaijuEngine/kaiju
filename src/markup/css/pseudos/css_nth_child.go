package pseudos

import (
	"errors"
	"fmt"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

func nth(args []string, count int) (int, int, error) {
	if len(args) == 0 {
		return 0, 0, errors.New("no arguments supplied")
	} else if count == 0 {
		return 0, 0, errors.New("no children")
	} else {
		start := 0
		skip := 0
		var err error
		if args[0] == "even" {
			args[0] = "2"
		} else if args[0] == "odd" {
			start = 1
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

func (p NthChild) Process(elm document.DocElement, value rules.SelectorPart) ([]document.DocElement, error) {
	if start, skip, err := nth(value.Args, len(elm.HTML.Children)); err == nil {
		selected := make([]document.DocElement, 0)
		for i := start; i < len(elm.HTML.Children); i += skip {
			selected = append(selected, *elm.HTML.Children[i].DocumentElement)
		}
		return selected, nil
	} else {
		return []document.DocElement{}, err
	}
}
