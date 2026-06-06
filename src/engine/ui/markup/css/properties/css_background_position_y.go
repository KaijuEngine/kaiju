/******************************************************************************/
/* css_background_position_y.go                                               */
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

func (p BackgroundPositionY) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected only one argument for background-position-y but got %d", len(values))
	}

	sd := panel.Base().ShaderData()
	if sd == nil {
		return nil
	}

	bgSizeY := panel.Background().Size().Y()
	y := helpers.NumFromLength(values[0].Str, host.Window)
	sd.UVs.SetY((bgSizeY-y)/bgSizeY - sd.UVs.W())
	return nil
}
