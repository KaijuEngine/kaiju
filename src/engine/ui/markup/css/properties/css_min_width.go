/******************************************************************************/
/* css_min_width.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (p MinWidth) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("MinWidth requires exactly 1 value")
	}

	if values[0].Str == "initial" {
		disableMinWidth(panel)
		return nil
	}

	minW := helpers.NumFromLength(values[0].Str, host.Window)
	if strings.HasSuffix(values[0].Str, "%") {
		layout := panel.Base().Layout()
		if layout.Ui().Entity().IsRoot() {
			minW = matrix.Float(host.Window.Width()) * minW
		} else if pUI := ui.FirstOnEntity(layout.Ui().Entity().Parent); pUI != nil {
			pLayout := pUI.Layout()
			s := pLayout.PixelSize().X() - pLayout.Padding().Horizontal() - pLayout.Border().Horizontal()
			if s < 0 {
				s = 0
			}
			minW = s * minW
		}
	}
	enableMinWidth(panel, minW)

	layout := panel.Base().Layout()
	layout.ScaleWidth(applyWidthConstraints(panel, layout.PixelSize().Width()))
	return nil
}
