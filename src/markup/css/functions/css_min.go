package functions

import (
	"errors"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (f Min) Process(panel *ui.Panel, elm document.DocElement, value rules.PropertyValue) (string, error) {
	return "", errors.New("not implemented")
}
