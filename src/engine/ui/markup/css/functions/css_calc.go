/******************************************************************************/
/* css_calc.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package functions

import (
	"strconv"
	"strings"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/rendering"
)

type calcOp int

const (
	calcOpNone = calcOp(iota)
	calcOpAdd
	calcOpSub
	calcOpMul
	calcOpDiv
)

type calcEntry struct {
	value float32
	op    calcOp
}

func (f Calc) Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error) {
	prop := value.Args[len(value.Args)-1]
	value.Args = value.Args[:len(value.Args)-1]
	entries := make([]calcEntry, len(value.Args))
	for i := range value.Args {
		switch value.Args[i] {
		case "+":
			entries[i].op = calcOpAdd
		case "-":
			entries[i].op = calcOpSub
		case "*":
			entries[i].op = calcOpMul
		case "/":
			entries[i].op = calcOpDiv
		default:
			v := helpers.NumFromLength(value.Args[i], panel.Base().Host().Window)
			if strings.HasSuffix(value.Args[i], "%") {
				switch prop {
				case "width":
					pl := elm.Parent.Value().UI.Layout()
					p := pl.Padding()
					v *= pl.PixelSize().Width() - p.X() - p.Z()
				case "height":
					pl := elm.Parent.Value().UI.Layout()
					p := pl.Padding()
					v *= pl.PixelSize().Height() - p.Y() - p.W()
				}
			}
			entries[i].value = v
		}
	}
	// Go through and do all the multiply and divide
	for i := range entries {
		if entries[i].op == calcOpMul {
			entries[i-1].value *= entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		} else if entries[i].op == calcOpDiv {
			entries[i-1].value /= entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		}
	}
	// Go through and do all the add and subtract
	for i := 0; i < len(entries); i++ {
		if entries[i].op == calcOpAdd {
			entries[i-1].value += entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		} else if entries[i].op == calcOpSub {
			entries[i-1].value -= entries[i+1].value
			entries = append(entries[:i], entries[i+2:]...)
			i--
		}
	}
	// Values above are already in scaled pixels; convert back to CSS px so the
	// result is not scaled a second time when it is parsed as a length.
	result := entries[0].value / rendering.UIScale
	return strconv.FormatFloat(float64(result), 'f', 5, 32) + "px", nil
}
