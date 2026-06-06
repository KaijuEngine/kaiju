/******************************************************************************/
/* css_flex_helpers.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"strconv"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
)

func parseFlexFloat(value string) (float32, bool) {
	f, err := strconv.ParseFloat(strings.TrimSpace(value), 32)
	return float32(f), err == nil
}

func setFlexBasis(panel *ui.Panel, value string, host *engine.Host) {
	value = strings.TrimSpace(value)
	layout := panel.Base().Layout()
	switch value {
	case "", "auto", "initial", "inherit", "unset":
		layout.SetFlexBasisAuto()
	default:
		layout.SetFlexBasis(helpers.NumFromLength(value, host.Window), strings.HasSuffix(value, "%"))
	}
}

func setFlexDirection(panel *ui.Panel, value string) bool {
	switch strings.TrimSpace(value) {
	case "row", "initial", "inherit", "unset":
		panel.SetFlexDirection(ui.FlexDirectionRow)
	case "row-reverse":
		panel.SetFlexDirection(ui.FlexDirectionRowReverse)
	case "column":
		panel.SetFlexDirection(ui.FlexDirectionColumn)
	case "column-reverse":
		panel.SetFlexDirection(ui.FlexDirectionColumnReverse)
	default:
		return false
	}
	return true
}

func setFlexWrap(panel *ui.Panel, value string) bool {
	switch strings.TrimSpace(value) {
	case "nowrap", "initial", "inherit", "unset":
		panel.SetFlexWrap(ui.FlexWrapNoWrap)
	case "wrap":
		panel.SetFlexWrap(ui.FlexWrapWrap)
	case "wrap-reverse":
		panel.SetFlexWrap(ui.FlexWrapWrapReverse)
	default:
		return false
	}
	return true
}

func parseFlexAlign(value string) (ui.FlexAlign, bool) {
	switch strings.TrimSpace(value) {
	case "auto":
		return ui.FlexAlignAuto, true
	case "normal", "stretch", "initial", "inherit", "unset":
		return ui.FlexAlignStretch, true
	case "start", "flex-start", "self-start", "baseline":
		return ui.FlexAlignStart, true
	case "end", "flex-end", "self-end":
		return ui.FlexAlignEnd, true
	case "center":
		return ui.FlexAlignCenter, true
	default:
		return ui.FlexAlignAuto, false
	}
}

func valuesToStrings(values []rules.PropertyValue) []string {
	out := make([]string, len(values))
	for i := range values {
		out[i] = values[i].Str
	}
	return out
}
