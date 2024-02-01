package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/helpers"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
	"kaiju/windowing"
	"strings"
)

var borderSizes = map[string]float32{
	"medium": 2,
	"thin":   1,
	"thick":  4,
}

func borderSizeFromStr(str string, window *windowing.Window, fallback float32) float32 {
	if val, ok := borderSizes[str]; ok {
		return val
	} else if strings.HasSuffix(str, "px") {
		return helpers.NumFromLength(str, window)
	} else {
		return fallback
	}
}

var borderStyleMap = map[string]ui.BorderStyle{
	"none":   ui.BorderStyleNone,
	"hidden": ui.BorderStyleHidden,
	"dotted": ui.BorderStyleDotted,
	"dashed": ui.BorderStyleDashed,
	"solid":  ui.BorderStyleSolid,
	"double": ui.BorderStyleDouble,
	"groove": ui.BorderStyleGroove,
	"ridge":  ui.BorderStyleRidge,
	"inset":  ui.BorderStyleInset,
	"outset": ui.BorderStyleOutset,
}

func borderStyleFromStr(str string, lrtb int, elm markup.DocElement) (ui.BorderStyle, bool) {
	if val, ok := borderStyleMap[str]; ok {
		return val, true
	} else if str == "initial" {
		// TODO:  Based on tag
		return ui.BorderStyleNone, true
	} else if str == "inherit" && elm.HTML.Parent != nil {
		return elm.HTML.Parent.DocumentElement.UI.(*ui.Panel).BorderStyle()[lrtb], true
	} else {
		return ui.BorderStyleNone, false
	}
}

// border-width border-style border-color|initial|inherit
func (p Border) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	BorderLeftWidth{}.Process(panel, elm, values[:1], host)
	BorderTopWidth{}.Process(panel, elm, values[:1], host)
	BorderRightWidth{}.Process(panel, elm, values[:1], host)
	BorderBottomWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderLeftStyle{}.Process(panel, elm, values[1:2], host)
		BorderTopStyle{}.Process(panel, elm, values[1:2], host)
		BorderRightStyle{}.Process(panel, elm, values[1:2], host)
		BorderBottomStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderLeftColor{}.Process(panel, elm, values[2:], host)
		BorderTopColor{}.Process(panel, elm, values[2:], host)
		BorderRightColor{}.Process(panel, elm, values[2:], host)
		BorderBottomColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
