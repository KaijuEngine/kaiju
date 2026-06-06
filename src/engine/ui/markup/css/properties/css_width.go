/******************************************************************************/
/* css_width.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/functions"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Width) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	if values[0].Str == "initial" {
		return nil
	}

	if values[0].Str == "fit-content" {
		panel.FitContentWidth()
		return nil
	}

	width := helpers.NumFromLength(values[0].Str, host.Window)

	panel.DontFitContentWidth()
	l := panel.Base().Layout()
	c := currentSizingConstraints(panel)
	if strings.HasSuffix(values[0].Str, "%") {
		if l.Ui().Entity().IsRoot() {
			finalW := applyWidthConstraints(panel, float32(host.Window.Width())*width)
			l.ScaleWidth(finalW)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleHeight(applyHeightConstraints(panel, finalW/c.AspectRatio))
			}
			return nil
		}
		pUI := ui.FirstOnEntity(l.Ui().Entity().Parent)
		if pUI != nil {
			parentPanel := pUI.ToPanel()
			if parentPanel.IsGrid() {
				// Child % width resolves to grid cell width (fixes div{ width: 100%; } in grid)
				cellW := parentPanel.GridCellWidth()
				finalW := applyWidthConstraints(panel, cellW*width)
				l.ScaleWidth(finalW)
				if c.HasAspectRatio() && c.AspectRatio > 0 {
					l.ScaleHeight(applyHeightConstraints(panel, finalW/c.AspectRatio))
				}
				return nil
			}
			pLayout := pUI.Layout()
			os := pLayout.PixelSize().X()
			s := os
			s -= pLayout.Padding().Horizontal()
			s -= pLayout.Border().Horizontal()
			if os > 0 && s < 0 {
				s = 0.001
			}
			finalW := applyWidthConstraints(panel, s*width)
			l.ScaleWidth(finalW)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleHeight(applyHeightConstraints(panel, finalW/c.AspectRatio))
			}
		}
	} else if values[0].IsFunction() {
		if values[0].Str == "calc" {
			val := values[0]
			val.Args = append(val.Args, "width")
			res, _ := functions.Calc{}.Process(panel, elm, val)
			width = helpers.NumFromLength(res, host.Window)
			width = applyWidthConstraints(panel, width)
			l.ScaleWidth(width)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleHeight(applyHeightConstraints(panel, width/c.AspectRatio))
			}
		}
	} else {
		width = applyWidthConstraints(panel, width)
		l.ScaleWidth(width)
		if c.HasAspectRatio() && c.AspectRatio > 0 {
			l.ScaleHeight(applyHeightConstraints(panel, width/c.AspectRatio))
		}
	}

	return nil
}
