package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// length|auto|initial|inherit
func (p MarginBottom) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("MarginBottom requires exactly 1 value")
	} else {
		current := panel.Layout().Margin()
		size := marginSizeFromStr(values[0].Str, host.Window)
		panel.Layout().SetMargin(current.X(), current.Y(), current.Z(), size)
		return nil
	}
}
