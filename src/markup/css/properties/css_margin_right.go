package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

// length|auto|initial|inherit
func (p MarginRight) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("MarginRight requires exactly 1 value")
	} else {
		current := panel.Layout().Margin()
		size := marginSizeFromStr(values[0].Str, host.Window)
		panel.Layout().SetMargin(current.X(), current.Y(), size, current.W())
		return nil
	}
}
