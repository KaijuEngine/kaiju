/******************************************************************************/
/* css_order.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"
	"strconv"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Order) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	if values[0].Str == "initial" || values[0].Str == "inherit" || values[0].Str == "unset" {
		panel.Base().Layout().SetFlexOrder(0)
		return nil
	}
	order, err := strconv.Atoi(values[0].Str)
	if err != nil {
		return fmt.Errorf("invalid order value %q", values[0].Str)
	}
	panel.Base().Layout().SetFlexOrder(order)
	return nil
}
