/******************************************************************************/
/* css_border.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/windowing"
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

func borderStyleFromStr(str string, lrtb int, elm *document.Element) (ui.BorderStyle, bool) {
	if val, ok := borderStyleMap[str]; ok {
		return val, true
	} else if str == "initial" {
		// TODO:  Based on tag
		return ui.BorderStyleNone, true
	} else if str == "inherit" && elm.Parent.Value() != nil {
		return elm.Parent.Value().UI.ToPanel().BorderStyle()[lrtb], true
	} else {
		return ui.BorderStyleNone, false
	}
}

func (Border) Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	// TODO:  The border args contain things like width, style, color
	//switch len(values) {
	//case 1:
	//	for i := range 3 {
	//		values = append(values, values[i])
	//	}
	//case 2:
	//	values = append(values, values[0])
	//	values = append(values, values[1])
	//case 3:
	//	values = append(values, values[1])
	//}
	//for i := 1; i < len(rules); i++ {
	//	removeRule := false
	//	switch rules[i].Property {
	//	case "border-top":
	//		values[0] = rules[i].Values[0]
	//		removeRule = true
	//	case "border-right":
	//		values[1] = rules[i].Values[0]
	//		removeRule = true
	//	case "border-bottom":
	//		values[2] = rules[i].Values[0]
	//		removeRule = true
	//	case "border-left":
	//		values[3] = rules[i].Values[0]
	//		removeRule = true
	//	}
	//	if removeRule {
	//		rules = slices.Delete(rules, i, i+1)
	//		i--
	//	}
	//}
	return values, rules
}

// border-width border-style border-color|initial|inherit
func (Border) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	if len(values) == 3 {
		BorderLeft{}.Process(panel, elm, values, host)
		BorderTop{}.Process(panel, elm, values, host)
		BorderRight{}.Process(panel, elm, values, host)
		BorderBottom{}.Process(panel, elm, values, host)
	}
	return nil
}
