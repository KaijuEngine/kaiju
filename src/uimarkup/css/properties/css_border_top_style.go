package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// none|hidden|dotted|dashed|solid|double|groove|ridge|inset|outset|initial|inherit
func (p BorderTopStyle) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderTopStyle requires 1 value")
	} else if border, ok := borderStyleFromStr(values[0].Str, 1, elm); !ok {
		return errors.New("BorderTopStyle: invalid border style")
	} else {
		borders := panel.BorderStyle()
		panel.SetBorderStyle(borders[0], border, borders[2], borders[3])
		return nil
	}
}
