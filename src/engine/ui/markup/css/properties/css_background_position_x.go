/******************************************************************************/
/* css_background_position_x.go                                               */
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

func (p BackgroundPositionX) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected only one argument for background-position-x but got %d", len(values))
	}

	sd := panel.Base().ShaderData()
	if sd == nil {
		return nil
	}

	bgSize := panel.Background().Size()
	x := helpers.NumFromLength(values[0].Str, host.Window)
	sd.UVs.SetX(x / bgSize.X())
	return nil
}
