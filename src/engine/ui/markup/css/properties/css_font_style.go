/******************************************************************************/
/* css_font_style.go                                                          */
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

func setChildrenFontStyle(elm *document.Element, style string) {
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		lbl.SetFontStyle(style)
	} else if elm.UI.IsType(ui.ElementTypeInput) {
		elm.UI.ToInput().SetFontStyle(style)
	} else {
		for _, child := range elm.Children {
			setChildrenFontStyle(child, style)
		}
	}
}

// normal|italic|oblique|initial|inherit
func (p FontStyle) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("FontWeight requires exactly 1 value")
	}

	setChildrenFontStyle(elm, values[0].Str)
	return nil
}
