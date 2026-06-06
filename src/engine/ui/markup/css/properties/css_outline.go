/******************************************************************************/
/* css_outline.go                                                             */
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
	"kaijuengine.com/matrix"
)

func (p Outline) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Outline requires 1-3 values")
	}

	width := float32(2)
	offset := panel.OutlineOffset()
	color := matrix.ColorBlack()
	style := ui.BorderStyle(ui.BorderStyleSolid)
	hasDrawableValue := false
	for _, value := range values {
		str := value.Str
		if str == "initial" || str == "none" || str == "hidden" {
			panel.SetOutline(0, offset, matrix.ColorTransparent())
			return nil
		}
		if parsedStyle, ok := borderStyleFromStr(str, 0, elm); ok {
			style = parsedStyle
			hasDrawableValue = true
			continue
		}
		parsedWidth := borderSizeFromStr(str, host.Window, -1)
		if parsedWidth >= 0 {
			width = parsedWidth
			hasDrawableValue = true
			continue
		}
		if parsedColor, ok := outlineColorFromStr(str); ok {
			color = parsedColor
			hasDrawableValue = true
			continue
		}
		return errors.New("Invalid outline value")
	}
	if !hasDrawableValue || style == ui.BorderStyleNone || style == ui.BorderStyleHidden || width <= 0 {
		panel.SetOutline(0, offset, matrix.ColorTransparent())
		return nil
	}
	panel.SetOutline(width, offset, color)
	return nil
}

func outlineColorFromStr(str string) (matrix.Color, bool) {
	if str == "currentColor" || str == "invert" {
		return matrix.ColorBlack(), true
	}
	if mapped, ok := helpers.ColorMap[str]; ok {
		str = mapped
	}
	hex := strings.TrimPrefix(str, "#")
	if !isHexColorString(hex) {
		return matrix.ColorTransparent(), false
	}
	color, err := matrix.ColorFromHexString(str)
	return color, err == nil
}

func isHexColorString(str string) bool {
	switch len(str) {
	case 3, 6, 8:
	default:
		return false
	}
	for _, r := range str {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
