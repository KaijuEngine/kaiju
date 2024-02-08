package functions

import (
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
	"strconv"
	"strings"
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

func (f Calc) Process(panel *ui.Panel, elm document.DocElement, value rules.PropertyValue) (string, error) {
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
			v := helpers.NumFromLength(value.Args[i], panel.Host().Window)
			if strings.HasSuffix(value.Args[i], "%") {
				if prop == "width" {
					pl := elm.HTML.Parent.DocumentElement.UI.Layout()
					p := pl.Padding()
					v *= pl.PixelSize().Width() - p.X() - p.Z()
				} else if prop == "height" {
					pl := elm.HTML.Parent.DocumentElement.UI.Layout()
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
	return strconv.FormatFloat(float64(entries[0].value), 'f', 5, 32) + "px", nil
}
