package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func setChildrenFontWeight(elm markup.DocElement, weight string, host *engine.Host) {
	if elm.HTML.IsText() {
		lbl := elm.UI.(*ui.Label)
		lbl.SetFontWeight(weight)
	} else {
		for _, child := range elm.HTML.Children {
			setChildrenFontWeight(*child.DocumentElement, weight, host)
		}
	}
}

// normal|bold|bolder|lighter|number|initial|inherit
func (p FontWeight) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("FontWeight requires exactly 1 value")
	}
	setChildrenFontWeight(elm, values[0].Str, host)
	return nil
}
