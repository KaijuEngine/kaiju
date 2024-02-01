package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// border-width border-style border-color|initial|inherit
func (p BorderRight) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	BorderRightWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderRightStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderRightColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
