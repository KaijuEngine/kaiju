/******************************************************************************/
/* css_background_size.go                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p BackgroundSize) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	sd := panel.Base().ShaderData()
	if sd == nil {
		return nil
	}
	bgSize := panel.Background().Size()
	switch len(values) {
	case 1:
		switch values[0].Str {
		case "contain":
			fallthrough
		case "cover":
			fallthrough
		case "auto":
			return fmt.Errorf("background-size does not yet implement %s", values[0].Str)
		}
		size := helpers.NumFromLength(values[0].Str, host.Window)
		sd.UVs.SetZ(size / bgSize.X())
		sd.UVs.SetW(size / bgSize.Y())
	case 2:
		width := helpers.NumFromLength(values[0].Str, host.Window)
		height := helpers.NumFromLength(values[1].Str, host.Window)
		sd.UVs.SetZ(width / bgSize.X())
		sd.UVs.SetW(height / bgSize.Y())
		sd.Size2D.SetZ(width)
		sd.Size2D.SetW(height)
	default:
	}
	return nil
}
