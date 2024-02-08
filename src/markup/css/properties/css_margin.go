package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
	"kaiju/windowing"
)

func marginSizeFromStr(str string, window *windowing.Window) float32 {
	if val, ok := borderSizes[str]; ok {
		return val
	} else {
		return helpers.NumFromLength(str, window)
	}
}

func (p Margin) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	var err error
	if len(values) == 1 {
		err = MarginLeft{}.Process(panel, elm, values, host)
		err = MarginTop{}.Process(panel, elm, values, host)
		err = MarginRight{}.Process(panel, elm, values, host)
		err = MarginBottom{}.Process(panel, elm, values, host)
	} else if len(values) == 2 {
		err = MarginTop{}.Process(panel, elm, values[:1], host)
		err = MarginBottom{}.Process(panel, elm, values[:1], host)
		err = MarginLeft{}.Process(panel, elm, values[1:], host)
		err = MarginRight{}.Process(panel, elm, values[1:], host)
	} else if len(values) == 3 {
		err = MarginTop{}.Process(panel, elm, values[:1], host)
		err = MarginRight{}.Process(panel, elm, values[1:2], host)
		err = MarginLeft{}.Process(panel, elm, values[1:2], host)
		err = MarginBottom{}.Process(panel, elm, values[2:], host)
	} else if len(values) == 4 {
		err = MarginTop{}.Process(panel, elm, values[:1], host)
		err = MarginRight{}.Process(panel, elm, values[1:2], host)
		err = MarginBottom{}.Process(panel, elm, values[2:3], host)
		err = MarginLeft{}.Process(panel, elm, values[3:], host)
	} else {
		err = errors.New("Margin requires 1-4 values")
	}
	return err
}
