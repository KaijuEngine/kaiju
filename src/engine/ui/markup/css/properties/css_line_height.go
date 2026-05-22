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

func lineHeightFromStr(str string, lbl *ui.Label, host *engine.Host) float32 {
	str = strings.TrimSpace(str)
	if str == "normal" {
		return 0
	}
	if v, err := strconv.ParseFloat(str, 32); err == nil {
		return lbl.FontSize() * float32(v)
	}
	height := helpers.NumFromLengthWithFont(str, host.Window, lbl.FontSize())
	if strings.HasSuffix(str, "%") {
		height *= lbl.FontSize()
	}
	return height
}

func setChildrenLineHeight(elm *document.Element, size string, host *engine.Host) {
	if elm.Stylizer.HasRule("line-height") {
		return
	}
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		lbl.SetLineHeight(lineHeightFromStr(size, lbl, host))
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
	for _, child := range elm.Children {
		setChildrenLineHeight(child, values[0].Str, host)
	}
	return nil
}
