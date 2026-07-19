/******************************************************************************/
/* css_line_height.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"strconv"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (LineHeight) Sort() int { return 1 }

func lineHeightFromStr(str string, fontSize float32, host *engine.Host) float32 {
	str = strings.TrimSpace(str)
	if str == "normal" {
		return 0
	}
	if v, err := strconv.ParseFloat(str, 32); err == nil {
		return fontSize * float32(v)
	}
	height := helpers.NumFromLengthWithFont(str, host.Window, fontSize)
	if strings.HasSuffix(str, "%") {
		height *= fontSize
	}
	return height
}

func setChildrenLineHeight(elm *document.Element, size string, host *engine.Host) {
	if elm == nil {
		return
	}
	if elm.Stylizer.HasRule("line-height") {
		return
	}
	if elm.UI == nil {
		for _, child := range elm.Children {
			setChildrenLineHeight(child, size, host)
		}
		return
	}
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		lbl.SetLineHeight(lineHeightFromStr(size, lbl.FontSize(), host))
	} else if elm.UI.IsType(ui.ElementTypeInput) {
		input := elm.UI.ToInput()
		input.SetLineHeight(lineHeightFromStr(size, input.FontSize(), host))
	} else if elm.UI.IsType(ui.ElementTypeTextArea) {
		textarea := elm.UI.ToTextArea()
		textarea.SetLineHeight(lineHeightFromStr(size, textarea.FontSize(), host))
	} else {
		for _, child := range elm.Children {
			setChildrenLineHeight(child, size, host)
		}
	}
}

func (p LineHeight) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("LineHeight requires exactly 1 value")
	}
	if elm == nil {
		return nil
	}
	if elm.UI == nil {
		for _, child := range elm.Children {
			setChildrenLineHeight(child, values[0].Str, host)
		}
		return nil
	}
	if elm.UI.IsType(ui.ElementTypeInput) {
		input := elm.UI.ToInput()
		input.SetLineHeight(lineHeightFromStr(values[0].Str, input.FontSize(), host))
		return nil
	}
	if elm.UI.IsType(ui.ElementTypeTextArea) {
		textarea := elm.UI.ToTextArea()
		textarea.SetLineHeight(lineHeightFromStr(values[0].Str, textarea.FontSize(), host))
		return nil
	}
	for _, child := range elm.Children {
		setChildrenLineHeight(child, values[0].Str, host)
	}
	return nil
}
