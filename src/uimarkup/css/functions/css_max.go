package functions

import (
	"errors"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (f Max) Process(panel *ui.Panel, elm markup.DocElement, value rules.PropertyValue) (string, error) {
	return "", errors.New("not implemented")
}
