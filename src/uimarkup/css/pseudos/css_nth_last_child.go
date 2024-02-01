package pseudos

import (
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (p NthLastChild) Process(elm markup.DocElement, value rules.SelectorPart) ([]markup.DocElement, error) {
	if start, skip, err := nth(value.Args, len(elm.HTML.Children)); err == nil {
		selected := make([]markup.DocElement, 0)
		for i := len(elm.HTML.Children) - 1 - start; i >= 0; i -= skip {
			selected = append(selected, *elm.HTML.Children[i].DocumentElement)
		}
		return selected, nil
	} else {
		return []markup.DocElement{}, err
	}
}
