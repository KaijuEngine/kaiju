package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func setChildrenFontStyle(elm document.DocElement, weight string, host *engine.Host) {
	if elm.HTML.IsText() {
		lbl := elm.UI.(*ui.Label)
		lbl.SetFontStyle(weight)
	} else {
		for _, child := range elm.HTML.Children {
			setChildrenFontStyle(*child.DocumentElement, weight, host)
		}
	}
}

// normal|italic|oblique|initial|inherit
func (p FontStyle) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("FontWeight requires exactly 1 value")
	}
	setChildrenFontStyle(elm, values[0].Str, host)
	return nil
}
