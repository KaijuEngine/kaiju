package pseudos

import (
	"errors"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
)

func (p FirstChild) Process(elm document.DocElement, value rules.SelectorPart) ([]document.DocElement, error) {
	if len(elm.HTML.Children) == 0 {
		return []document.DocElement{}, errors.New("no children")
	} else {
		return []document.DocElement{*elm.HTML.Children[0].DocumentElement}, nil
	}
}
