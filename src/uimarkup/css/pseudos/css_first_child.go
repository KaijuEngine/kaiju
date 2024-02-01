package pseudos

import (
	"errors"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (p FirstChild) Process(elm markup.DocElement, value rules.SelectorPart) ([]markup.DocElement, error) {
	if len(elm.HTML.Children) == 0 {
		return []markup.DocElement{}, errors.New("no children")
	} else {
		return []markup.DocElement{*elm.HTML.Children[0].DocumentElement}, nil
	}
}
