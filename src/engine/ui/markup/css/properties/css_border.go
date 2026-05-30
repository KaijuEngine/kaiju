/******************************************************************************/
/* css_border.go                                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"slices"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/windowing"
)

const mergedBorderSidesSentinel = "__kaiju_merged_border_sides__"

const (
	borderSideLeft = iota
	borderSideTop
	borderSideRight
	borderSideBottom
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

func (Border) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	sides := [4][]rules.PropertyValue{
		cloneBorderValues(values),
		cloneBorderValues(values),
		cloneBorderValues(values),
		cloneBorderValues(values),
	}
	merged := false
	for i := 1; i < len(ruleList); i++ {
		sideIdx := -1
		switch ruleList[i].Property {
		case "border":
			return values, ruleList[1:]
		case "border-left":
			sideIdx = borderSideLeft
		case "border-top":
			sideIdx = borderSideTop
		case "border-right":
			sideIdx = borderSideRight
		case "border-bottom":
			sideIdx = borderSideBottom
		case "border-width":
			for side := range sides {
				if width, ok := borderWidthValueForSide(ruleList[i].Values, side); ok {
					sides[side] = setBorderSideWidth(sides[side], width)
					merged = true
				}
			}
			ruleList = slices.Delete(ruleList, i, i+1)
			i--
			continue
		case "border-left-width":
			sideIdx = borderSideLeft
		case "border-top-width":
			sideIdx = borderSideTop
		case "border-right-width":
			sideIdx = borderSideRight
		case "border-bottom-width":
			sideIdx = borderSideBottom
		}
		if sideIdx >= 0 {
			if isBorderWidthRule(ruleList[i].Property) {
				if width, ok := firstBorderValue(ruleList[i].Values); ok {
					sides[sideIdx] = setBorderSideWidth(sides[sideIdx], width)
				}
			} else {
				sides[sideIdx] = mergeBorderValues(sides[sideIdx], ruleList[i].Values)
			}
			ruleList = slices.Delete(ruleList, i, i+1)
			merged = true
			i--
		}
	}
	if merged {
		values = []rules.PropertyValue{{Str: mergedBorderSidesSentinel}}
		for i := range sides {
			values = append(values, rules.PropertyValue{Num: float32(len(sides[i]))})
			values = append(values, sides[i]...)
		}
		ruleList[0].Values = values
	}
	return values, ruleList
}

// border-width border-style border-color|initial|inherit
func (Border) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) > 0 && values[0].Str == mergedBorderSidesSentinel {
		i := 1
		sideRules := []struct {
			name    string
			process func(*ui.Panel, *document.Element, []rules.PropertyValue, *engine.Host) error
		}{
			{"border-left", BorderLeft{}.Process},
			{"border-top", BorderTop{}.Process},
			{"border-right", BorderRight{}.Process},
			{"border-bottom", BorderBottom{}.Process},
		}
		for _, side := range sideRules {
			if i >= len(values) {
				return errors.New("merged border data is missing side values")
			}
			count := int(values[i].Num)
			i++
			if count <= 0 || i+count > len(values) {
				return errors.New("merged border data has invalid side values")
			}
			if err := side.process(panel, elm, values[i:i+count], host); err != nil {
				return err
			}
			i += count
		}
		return nil
	}
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

func cloneBorderValues(values []rules.PropertyValue) []rules.PropertyValue {
	out := make([]rules.PropertyValue, len(values))
	for i := range values {
		out[i] = values[i].Clone()
	}
	return out
}

func mergeBorderValues(base, override []rules.PropertyValue) []rules.PropertyValue {
	out := cloneBorderValues(base)
	for i := range override {
		if i < len(out) {
			out[i] = override[i].Clone()
		} else {
			out = append(out, override[i].Clone())
		}
	}
	return out
}

func isBorderWidthRule(property string) bool {
	switch property {
	case "border-left-width", "border-top-width", "border-right-width", "border-bottom-width":
		return true
	default:
		return false
	}
}

func borderWidthValueForSide(values []rules.PropertyValue, side int) (rules.PropertyValue, bool) {
	values = expandFourSideValues(values)
	if len(values) != 4 {
		return rules.PropertyValue{}, false
	}
	switch side {
	case borderSideLeft:
		return values[3], true
	case borderSideTop:
		return values[0], true
	case borderSideRight:
		return values[1], true
	case borderSideBottom:
		return values[2], true
	default:
		return rules.PropertyValue{}, false
	}
}

func setBorderSideWidth(values []rules.PropertyValue, width rules.PropertyValue) []rules.PropertyValue {
	out := cloneBorderValues(values)
	if len(out) == 0 {
		return []rules.PropertyValue{width.Clone()}
	}
	out[0] = width.Clone()
	return out
}

func mergeFutureBorderSideWidths(values []rules.PropertyValue, ruleList []rules.Rule, side int, sideProperty string) ([]rules.PropertyValue, []rules.Rule) {
	for i := 1; i < len(ruleList); i++ {
		switch ruleList[i].Property {
		case "border", sideProperty:
			return values, ruleList[1:]
		case "border-width":
			if width, ok := borderWidthValueForSide(ruleList[i].Values, side); ok {
				values = setBorderSideWidth(values, width)
			}
		case sideProperty + "-width":
			if width, ok := firstBorderValue(ruleList[i].Values); ok {
				values = setBorderSideWidth(values, width)
				ruleList = slices.Delete(ruleList, i, i+1)
				i--
			}
		}
	}
	ruleList[0].Values = values
	return values, ruleList
}

func preprocBorderSideWidth(values []rules.PropertyValue, ruleList []rules.Rule, sideProperty string) ([]rules.PropertyValue, []rules.Rule) {
	for i := 1; i < len(ruleList); i++ {
		switch ruleList[i].Property {
		case "border", "border-width", sideProperty, sideProperty + "-width":
			return values, ruleList[1:]
		}
	}
	return values, ruleList
}

func firstBorderValue(values []rules.PropertyValue) (rules.PropertyValue, bool) {
	if len(values) == 0 {
		return rules.PropertyValue{}, false
	}
	return values[0], true
}
