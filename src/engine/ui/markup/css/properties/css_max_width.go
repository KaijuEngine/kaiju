/******************************************************************************/
/* css_max_width.go                                                           */
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

func (p MaxWidth) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("MaxWidth requires exactly 1 value")
	}
	if values[0].Str == "initial" {
		disableMaxWidth(panel)
		return nil
	}
	maxW := helpers.NumFromLength(values[0].Str, host.Window)
	if strings.HasSuffix(values[0].Str, "%") {
		layout := panel.Base().Layout()
		if layout.Ui().Entity().IsRoot() {
			maxW = float32(host.Window.Width()) * maxW
		} else if pUI := ui.FirstOnEntity(layout.Ui().Entity().Parent); pUI != nil {
			pLayout := pUI.Layout()
			s := pLayout.PixelSize().X() - pLayout.Padding().Horizontal() - pLayout.Border().Horizontal()
			if s < 0 {
				s = 0
			}
			maxW = s * maxW
		}
	}
	enableMaxWidth(panel, maxW)
	layout := panel.Base().Layout()
	layout.ScaleWidth(applyWidthConstraints(panel, layout.PixelSize().Width()))
	return nil
}
