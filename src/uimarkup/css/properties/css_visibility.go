package properties

import (
	"errors"
	"fmt"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// visibility: visible|hidden|collapse|initial|inherit;
func (p Visibility) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Visibility arguments expects 1 argument only but received: %d", len(values))
	}

	s := values[0].Str
	switch s {
	case "initial":
		fallthrough
	case "visible":
		panel.Entity().Activate()
	case "hidden":
		panel.Entity().Deactivate()
	case "collapse": //Meant for table rows (<tr>), row groups (<tbody>), columns (<col>), column groups (<colgroup>)
		return errors.New("Not implemented [collapse]")
	case "inherit":
		panel.Entity().SetActive(panel.Entity().Parent.IsActive())
	}

	return nil
}
