package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

// border-width border-style border-color|initial|inherit
func (p BorderBottom) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	BorderBottomWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderBottomStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderBottomColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
