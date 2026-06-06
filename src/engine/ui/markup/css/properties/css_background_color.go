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
			// A label's font now blends against its calculated surface (the
			// real opaque color behind it, resolved by walking up the panel
			// tree), so "inherit" needs no work: leaving the label background
			// transparent makes the text track whatever the parent renders.
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
	}
	// Child text no longer needs its background pushed down: each label resolves
	// its own opaque surface by walking up the panel tree (CalculatedBGColor),
	// which composites partial transparency correctly at every level.
	return nil
}
