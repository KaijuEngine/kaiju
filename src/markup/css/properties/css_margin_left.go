package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

// length|auto|initial|inherit
func (p MarginLeft) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("MarginLeft requires exactly 1 value")
	} else {
		current := panel.Layout().Margin()
		size := marginSizeFromStr(values[0].Str, host.Window)
		panel.Layout().SetMargin(size, current.Y(), current.Z(), current.W())
		return nil
	}
}
