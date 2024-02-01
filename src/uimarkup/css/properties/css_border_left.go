package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// border-width border-style border-color|initial|inherit
func (p BorderLeft) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	BorderLeftWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderLeftStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderLeftColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
