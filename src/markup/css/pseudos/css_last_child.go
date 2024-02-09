package pseudos

import (
	"errors"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

func (p LastChild) Process(elm document.DocElement, value rules.SelectorPart) ([]document.DocElement, error) {
	if len(elm.HTML.Children) == 0 {
		return []document.DocElement{}, errors.New("no children")
	} else {
		idx := len(elm.HTML.Children) - 1
		return []document.DocElement{*elm.HTML.Children[idx].DocumentElement}, nil
	}
}
