package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
	"strconv"
)

func (p ZIndex) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("ZIndex requires exactly 1 value")
	} else {
		val, _ := strconv.ParseFloat(values[0].Str, 64)
		z := float32(val)
		p := elm.HTML.Parent
		for p != nil && p.DocumentElement != nil {
			z += p.DocumentElement.UI.Layout().Z()
			p = p.Parent
		}
		panel.Layout().SetZ(-z)
		return nil
	}
}
