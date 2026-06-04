/******************************************************************************/
/* css_height.go                                                              */
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

func (p Height) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	if values[0].Str == "initial" {
		return nil
	}

	if values[0].Str == "fit-content" {
		panel.FitContentHeight()
		return nil
	}

	height := helpers.NumFromLength(values[0].Str, host.Window)

	panel.DontFitContentHeight()
	l := panel.Base().Layout()
	c := currentSizingConstraints(panel)
	if strings.HasSuffix(values[0].Str, "%") {
		if l.Ui().Entity().IsRoot() {
			finalH := applyHeightConstraints(panel, float32(host.Window.Height())*height)
			l.ScaleHeight(finalH)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleWidth(applyWidthConstraints(panel, finalH*c.AspectRatio))
			}
			return nil
		}
		pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
		s := pLayout.PixelSize().Y()
		s -= pLayout.Padding().Vertical()
		s -= pLayout.Border().Vertical()
		finalH := applyHeightConstraints(panel, s*height)
		l.ScaleHeight(finalH)
		if c.HasAspectRatio() && c.AspectRatio > 0 {
			l.ScaleWidth(applyWidthConstraints(panel, finalH*c.AspectRatio))
		}
	} else if values[0].IsFunction() {
		if values[0].Str == "calc" {
			val := values[0]
			val.Args = append(val.Args, "height")
			res, _ := functions.Calc{}.Process(panel, elm, val)
			height = helpers.NumFromLength(res, host.Window)
			height = applyHeightConstraints(panel, height)
			l.ScaleHeight(height)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleWidth(applyWidthConstraints(panel, height*c.AspectRatio))
			}
		}
	} else {
		height = applyHeightConstraints(panel, height)
		l.ScaleHeight(height)
		if c.HasAspectRatio() && c.AspectRatio > 0 {
			l.ScaleWidth(applyWidthConstraints(panel, height*c.AspectRatio))
		}
	}

	return nil
}
