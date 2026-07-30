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
	"kaijuengine.com/matrix"
)

func fontSizeFromStr(size string, host *engine.Host, emSize matrix.Float) matrix.Float {
	return helpers.NumFromLengthWithFont(size, host.Window, emSize)
}

func setChildrenFontSize(elm *document.Element, size string, host *engine.Host) {
	if elm.Stylizer.HasRule("font-size") {
		return
	}
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		lbl.SetFontSize(fontSizeFromStr(size, host, host.FontCache().EMSize(lbl.FontFace())))
	} else if elm.UI.IsType(ui.ElementTypeInput) {
		input := elm.UI.ToInput()
		input.SetFontSize(fontSizeFromStr(size, host, host.FontCache().EMSize(input.FontFace())))
	} else if elm.UI.IsType(ui.ElementTypeTextArea) {
		textarea := elm.UI.ToTextArea()
		textarea.SetFontSize(fontSizeFromStr(size, host, host.FontCache().EMSize(textarea.FontFace())))
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
	if elm.UI.IsType(ui.ElementTypeInput) {
		input := elm.UI.ToInput()
		input.SetFontSize(fontSizeFromStr(values[0].Str, host, host.FontCache().EMSize(input.FontFace())))
		return nil
	}
	if elm.UI.IsType(ui.ElementTypeTextArea) {
		textarea := elm.UI.ToTextArea()
		textarea.SetFontSize(fontSizeFromStr(values[0].Str, host, host.FontCache().EMSize(textarea.FontFace())))
		return nil
	}
	for _, child := range elm.Children {
		setChildrenFontSize(child, values[0].Str, host)
	}
	return nil
}
