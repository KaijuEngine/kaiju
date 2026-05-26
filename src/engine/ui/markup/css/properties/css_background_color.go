/******************************************************************************/
/* css_background_color.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/functions"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func setChildTextBackgroundColor(elm *document.Element, color matrix.Color) {
	for _, c := range elm.Children {
		if c.IsText() {
			c.UI.ToLabel().SetBGColor(color)
		} else if !c.UI.IsType(ui.ElementTypeLabel) && c.UI.ToPanel().Background() == nil { // Only continue if transparent
			setChildTextBackgroundColor(c, color)
		}
	}
}

func (p BackgroundColor) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	}

	isLabel := elm.UI.IsType(ui.ElementTypeLabel)
	// Images used for background are not colored
	applyPanelColor := false
	if !isLabel {
		bg := elm.UI.ToPanel().Background()
		applyPanelColor = bg == nil || bg.Key == assets.TextureSquare
	}

	hex := values[0].Str
	if hex == "inherit" {
		if isLabel {
			pBase := panel.Base()
			elm.UI.AddEvent(ui.EventTypeRender, func() {
				if pBase.Entity().Parent != nil {
					p := ui.FirstPanelOnEntity(pBase.Entity().Parent)
					elm.UI.ToLabel().SetBGColor(p.Base().ShaderData().FgColor)
				}
			})
			return nil
		}
		if applyPanelColor {
			pBase := panel.Base()
			pBase.AddEvent(ui.EventTypeRender, func() {
				if pBase.Entity().Parent != nil {
					p := ui.FirstPanelOnEntity(pBase.Entity().Parent)
					panel.SetColor(p.Base().ShaderData().FgColor)
				}
			})
		}
		return nil
	}

	switch values[0].Str {
	case "rgb":
		hex, _ = functions.Rgb{}.Process(panel, elm, values[0])
	case "rgba":
		hex, _ = functions.Rgba{}.Process(panel, elm, values[0])
	}

	if newHex, ok := helpers.ColorMap[hex]; ok {
		hex = newHex
	}

	color, err := matrix.ColorFromHexString(hex)
	if err != nil {
		return err
	}

	if isLabel {
		elm.UI.ToLabel().SetBGColor(color)
		return nil
	}
	if applyPanelColor || panel.Base().Type() == ui.ElementTypeImage {
		panel.SetColor(color)
	}
	if panel.Base().IsType(ui.ElementTypeInput) {
		panel.Base().ToInput().SetBGColor(color)
	} else if panel.Base().IsType(ui.ElementTypeTextArea) {
		panel.Base().ToTextArea().SetBGColor(color)
	} else if !panel.HasEnforcedColor() {
		setChildTextBackgroundColor(elm, color)
	}
	return nil
}
