/******************************************************************************/
/* css_font_size.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func setChildrenFontSize(elm *document.Element, size string, host *engine.Host) {
	if elm.Stylizer.HasRule("font-size") {
		return
	}
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		size := helpers.NumFromLengthWithFont(size, host.Window,
			host.FontCache().EMSize(lbl.FontFace()))
		lbl.SetFontSize(size)
	} else {
		for _, child := range elm.Children {
			setChildrenFontSize(child, size, host)
		}
	}
}

func (p FontSize) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("FontSize requires exactly 1 value")
	}
	for _, child := range elm.Children {
		setChildrenFontSize(child, values[0].Str, host)
	}
	return nil
}
