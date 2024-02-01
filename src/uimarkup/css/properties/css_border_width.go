package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/helpers"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

func (p BorderWidth) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	b := [4]float32{}
	if len(values) == 1 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = b[0]
		b[2] = b[0]
		b[3] = b[0]
	} else if len(values) == 2 {
		// Top/bottom left/right
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[0] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = b[1]
	} else if len(values) == 3 {
		// Top left/right bottom
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[0] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = helpers.NumFromLength(values[2].Str, host.Window)
	} else if len(values) == 4 {
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[2] = helpers.NumFromLength(values[1].Str, host.Window)
		b[3] = helpers.NumFromLength(values[2].Str, host.Window)
		b[0] = helpers.NumFromLength(values[3].Str, host.Window)
	} else {
		return errors.New("Invalid number of values for BorderRadius")
	}
	panel.SetBorderSize(b[0], b[1], b[2], b[3])
	return nil
}
