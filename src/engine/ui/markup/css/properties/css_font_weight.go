/******************************************************************************/
/* css_font_weight.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func setChildrenFontWeight(elm *document.Element, weight string) {
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		lbl.SetFontWeight(weight)
	} else if elm.UI.IsType(ui.ElementTypeInput) {
		elm.UI.ToInput().SetFontWeight(weight)
	} else if elm.UI.IsType(ui.ElementTypeTextArea) {
		elm.UI.ToTextArea().SetFontWeight(weight)
	} else {
		for _, child := range elm.Children {
			setChildrenFontWeight(child, weight)
		}
	}
}

// normal|bold|bolder|lighter|number|initial|inherit
func (p FontWeight) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("FontWeight requires exactly 1 value")
	}

	setChildrenFontWeight(elm, values[0].Str)
	return nil
}
