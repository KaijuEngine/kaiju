package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// none|hidden|dotted|dashed|solid|double|groove|ridge|inset|outset|initial|inherit
func (p BorderLeftStyle) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderLeftStyle requires 1 value")
	} else if border, ok := borderStyleFromStr(values[0].Str, 0, elm); !ok {
		return errors.New("BorderLeftStyle: invalid border style")
	} else {
		borders := panel.BorderStyle()
		panel.SetBorderStyle(border, borders[1], borders[2], borders[3])
		return nil
	}
}
