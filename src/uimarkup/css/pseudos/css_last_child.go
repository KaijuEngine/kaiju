package pseudos

import (
	"errors"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (p LastChild) Process(elm markup.DocElement, value rules.SelectorPart) ([]markup.DocElement, error) {
	if len(elm.HTML.Children) == 0 {
		return []markup.DocElement{}, errors.New("no children")
	} else {
		idx := len(elm.HTML.Children) - 1
		return []markup.DocElement{*elm.HTML.Children[idx].DocumentElement}, nil
	}
}
