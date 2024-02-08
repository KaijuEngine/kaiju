package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

// none|hidden|dotted|dashed|solid|double|groove|ridge|inset|outset|initial|inherit
func (p BorderBottomStyle) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderBottomStyle requires 1 value")
	} else if border, ok := borderStyleFromStr(values[0].Str, 3, elm); !ok {
		return errors.New("BorderBottomStyle: invalid border style")
	} else {
		borders := panel.BorderStyle()
		panel.SetBorderStyle(borders[0], borders[1], borders[2], border)
		return nil
	}
}
