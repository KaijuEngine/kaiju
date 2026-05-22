/******************************************************************************/
/* css_min_height.go                                                          */
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
)

func (p MinHeight) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("MinHeight requires exactly 1 value")
	}

	if values[0].Str == "initial" {
		disableMinHeight(panel)
		return nil
	}

	minH := helpers.NumFromLength(values[0].Str, host.Window)
	if strings.HasSuffix(values[0].Str, "%") {
		layout := panel.Base().Layout()
		if layout.Ui().Entity().IsRoot() {
			minH = float32(host.Window.Height()) * minH
		} else if pUI := ui.FirstOnEntity(layout.Ui().Entity().Parent); pUI != nil {
			pLayout := pUI.Layout()
			s := pLayout.PixelSize().Y() - pLayout.Padding().Vertical() - pLayout.Border().Vertical()
			if s < 0 {
				s = 0
			}
			minH = s * minH
		}
	}
	enableMinHeight(panel, minH)

	layout := panel.Base().Layout()
	layout.ScaleHeight(applyHeightConstraints(panel, layout.PixelSize().Height()))
	return nil
}
