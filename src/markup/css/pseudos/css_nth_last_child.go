package pseudos

import (
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

func (p NthLastChild) Process(elm document.DocElement, value rules.SelectorPart) ([]document.DocElement, error) {
	if start, skip, err := nth(value.Args, len(elm.HTML.Children)); err == nil {
		selected := make([]document.DocElement, 0)
		for i := len(elm.HTML.Children) - 1 - start; i >= 0; i -= skip {
			selected = append(selected, *elm.HTML.Children[i].DocumentElement)
		}
		return selected, nil
	} else {
		return []document.DocElement{}, err
	}
}
