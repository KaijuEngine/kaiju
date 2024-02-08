package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
	"kaiju/windowing"
)

func paddingSizeFromString(elm document.DocElement, str string, idx matrix.VectorComponent, window *windowing.Window) (matrix.Vec4, error) {
	current := elm.UI.Layout().Padding()
	size := current[idx]
	if str == "initial" {
		size = 0
	} else if str == "inherit" {
		if elm.HTML.Parent == nil {
			size = 0
		} else {
			size = elm.HTML.Parent.DocumentElement.UI.Layout().Padding()[idx]
		}
	} else {
		size = helpers.NumFromLength(str, window)
	}
	current[idx] = size
	return current, nil
}

// length|initial|inherit
func (p Padding) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	var err error
	if len(values) == 1 {
		// all
		err = PaddingLeft{}.Process(panel, elm, values, host)
		err = PaddingRight{}.Process(panel, elm, values, host)
		err = PaddingTop{}.Process(panel, elm, values, host)
		err = PaddingBottom{}.Process(panel, elm, values, host)
	} else if len(values) == 2 {
		// top/bottom, left/right
		err = PaddingTop{}.Process(panel, elm, values[:1], host)
		err = PaddingBottom{}.Process(panel, elm, values[:1], host)
		err = PaddingLeft{}.Process(panel, elm, values[1:], host)
		err = PaddingRight{}.Process(panel, elm, values[1:], host)
	} else if len(values) == 3 {
		// top, left/right, bottom
		err = PaddingTop{}.Process(panel, elm, values[:1], host)
		err = PaddingLeft{}.Process(panel, elm, values[1:2], host)
		err = PaddingRight{}.Process(panel, elm, values[1:2], host)
		err = PaddingBottom{}.Process(panel, elm, values[2:], host)
	} else if len(values) == 4 {
		// top, right, bottom, left
		err = PaddingTop{}.Process(panel, elm, values[:1], host)
		err = PaddingRight{}.Process(panel, elm, values[1:2], host)
		err = PaddingBottom{}.Process(panel, elm, values[2:3], host)
		err = PaddingLeft{}.Process(panel, elm, values[3:], host)
	} else {
		err = errors.New("Padding: Expecting 1-4 values")
	}
	return err
}
