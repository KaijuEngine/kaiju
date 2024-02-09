package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

// none|hidden|dotted|dashed|solid|double|groove|ridge|inset|outset|initial|inherit
func (p BorderStyle) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	problems := []error{errors.New("BorderStyle not implemented")}

	return problems[0]
}
