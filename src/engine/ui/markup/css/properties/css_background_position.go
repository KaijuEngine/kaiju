/******************************************************************************/
/* css_background_position.go                                                 */
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

func (p BackgroundPosition) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	switch len(values) {
	case 1:
		if err := (BackgroundPositionX{}.Process(panel, elm, values, host)); err != nil {
			return err
		}
		return BackgroundPositionY{}.Process(panel, elm, values, host)
	case 2:
		if err := (BackgroundPositionX{}.Process(panel, elm, values[0:1], host)); err != nil {
			return err
		}
		return BackgroundPositionY{}.Process(panel, elm, values[1:], host)
	default:
		return fmt.Errorf("invalid number of arguments to CSS background-position, expected %d got %d", 2, len(values))
	}
}
