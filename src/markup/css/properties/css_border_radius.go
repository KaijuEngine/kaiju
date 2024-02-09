package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (p BorderRadius) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	b := [4]float32{}
	if len(values) == 1 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = b[0]
		b[2] = b[0]
		b[3] = b[0]
	} else if len(values) == 2 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = b[0]
		b[2] = helpers.NumFromLength(values[1].Str, host.Window)
		b[3] = b[2]
	} else if len(values) == 3 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = helpers.NumFromLength(values[2].Str, host.Window)
		b[3] = b[1]
	} else if len(values) == 4 {
		for i := 0; i < 4; i++ {
			b[i] = helpers.NumFromLength(values[i].Str, host.Window)
		}
	} else {
		return errors.New("Invalid number of values for BorderRadius")
	}
	panel.SetBorderRadius(b[0], b[1], b[2], b[3])
	return nil
}
