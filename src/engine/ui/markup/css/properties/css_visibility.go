/******************************************************************************/
/* css_visibility.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

// visibility: visible|hidden|collapse|initial|inherit;
func (p Visibility) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Visibility arguments expects 1 argument only but received: %d", len(values))
	}

	s := values[0].Str
	switch s {
	case "initial":
		fallthrough
	case "visible":
		panel.Base().Entity().Activate()
	case "hidden":
		panel.Base().Entity().Deactivate()
	case "collapse": //Meant for table rows (<tr>), row groups (<tbody>), columns (<col>), column groups (<colgroup>)
		return errors.New("Not implemented [collapse]")
	case "inherit":
		panel.Base().Entity().SetActive(panel.Base().Entity().Parent.IsActive())
	}

	return nil
}
