/******************************************************************************/
/* css_padding.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"slices"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/windowing"
)

func paddingSizeFromString(elm *document.Element, str string, idx matrix.VectorComponent, window *windowing.Window) (matrix.Vec4, error) {
	current := elm.UI.Layout().Padding()
	size := current[idx]
	switch str {
	case "initial":
		size = 0
	case "inherit":
		if elm.Parent.Value() == nil {
			size = 0
		} else {
			size = elm.Parent.Value().UI.Layout().Padding()[idx]
		}
	default:
		size = helpers.NumFromLength(str, window)
	}
	current[idx] = size
	return current, nil
}

func (Padding) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	switch len(values) {
	case 1:
		for i := range 3 {
			values = append(values, values[i])
		}
	case 2:
		values = append(values, values[0])
		values = append(values, values[1])
	case 3:
		values = append(values, values[1])
	}
	for i := 1; i < len(ruleList); i++ {
		removeRule := false
		switch ruleList[i].Property {
		case "padding-top":
			values[0] = ruleList[i].Values[0]
			removeRule = true
		case "padding-right":
			values[1] = ruleList[i].Values[0]
			removeRule = true
		case "padding-bottom":
			values[2] = ruleList[i].Values[0]
			removeRule = true
		case "padding-left":
			values[3] = ruleList[i].Values[0]
			removeRule = true
		}
		if removeRule {
			ruleList = slices.Delete(ruleList, i, i+1)
			i--
		}
	}
	ruleList[0].Values = values
	return values, ruleList
}

// length|initial|inherit
func (Padding) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
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
