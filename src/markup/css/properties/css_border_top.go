package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

// border-width border-style border-color|initial|inherit
func (p BorderTop) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("BorderTop requires 1-3 values")
	}
	BorderTopWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderTopStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderTopColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
