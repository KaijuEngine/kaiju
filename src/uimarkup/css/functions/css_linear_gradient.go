package functions

import (
	"errors"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (f LinearGradient) Process(panel *ui.Panel, elm markup.DocElement, value rules.PropertyValue) (string, error) {
	return "", errors.New("not implemented")
}
