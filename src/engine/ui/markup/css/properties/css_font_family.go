/******************************************************************************/
/* css_font_family.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"fmt"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

func setChildrenFontFace(elm *document.Element, face rendering.FontFace) {
	defer tracing.NewRegion("properties.setChildrenFontFace").End()
	if elm == nil {
		return
	}
	if elm.UI == nil {
		for _, child := range elm.Children {
			setChildrenFontFace(child, face)
		}
		return
	}
	if elm.IsText() {
		lbl := elm.UI.ToLabel()
		lbl.SetFontFace(face)
	} else if elm.UI.IsType(ui.ElementTypeInput) {
		elm.UI.ToInput().SetFontFace(face)
	} else if elm.UI.IsType(ui.ElementTypeTextArea) {
		elm.UI.ToTextArea().SetFontFace(face)
	} else {
		for _, child := range elm.Children {
			setChildrenFontFace(child, face)
		}
	}
}

func (p FontFamily) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	defer tracing.NewRegion("FontFamily.Process").End()
	// TODO:  Support monospace
	if values[0].Str[0] != '\'' && values[0].Str[0] != '"' {
		return fmt.Errorf("expected first value to be a string of the font name, but was: %s", values[0].Str)
	}
	faceName := strings.TrimSpace(strings.Trim(strings.Trim(values[0].Str, "'"), `"`))
	if faceName == "" {
		return errors.New("the font face supplied to CSS font-family was blank")
	}
	setChildrenFontFace(elm, rendering.FontFace(faceName))
	return nil
}
