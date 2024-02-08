package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (p BackgroundPositionX) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	problems := []error{errors.New("BackgroundPositionX not implemented")}

	return problems[0]
}