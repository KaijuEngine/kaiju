package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// none|hidden|dotted|dashed|solid|double|groove|ridge|inset|outset|initial|inherit
func (p BorderRightStyle) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderRightStyle requires 1 value")
	} else if border, ok := borderStyleFromStr(values[0].Str, 2, elm); !ok {
		return errors.New("BorderRightStyle: invalid border style")
	} else {
		borders := panel.BorderStyle()
		panel.SetBorderStyle(borders[0], borders[1], border, borders[3])
		return nil
	}
}
