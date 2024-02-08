package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func setChildrenLineHeight(elm document.DocElement, size string, host *engine.Host) {
	if elm.HTML.IsText() {
		lbl := elm.UI.(*ui.Label)
		size := helpers.NumFromLengthWithFont(size, host.Window,
			host.FontCache().EMSize(lbl.FontFace()))
		lbl.SetLineHeight(size)
	} else {
		for _, child := range elm.HTML.Children {
			setChildrenLineHeight(*child.DocumentElement, size, host)
		}
	}
}

func (p LineHeight) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("LineHeight requires exactly 1 value")
	}
	setChildrenLineHeight(elm, values[0].Str, host)
	return nil
}
