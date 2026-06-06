/******************************************************************************/
/* css_aspect_ratio.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p AspectRatio) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return fmt.Errorf("AspectRatio requires at least 1 value")
	}

	valueStrs := make([]string, 0, len(values))
	for i := range values {
		valueStrs = append(valueStrs, values[i].Str)
	}
	ratio, ok := parseRatio(valueStrs)
	if !ok {
		disableAspectRatio(panel)
		return nil
	}
	enableAspectRatio(panel, ratio)

	layout := panel.Base().Layout()
	width := applyWidthConstraints(panel, layout.PixelSize().Width())
	height := applyHeightConstraints(panel, width/ratio)
	layout.Scale(width, height)
	return nil
}
