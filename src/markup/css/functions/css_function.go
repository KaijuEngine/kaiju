package functions

import (
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

type Function interface {
	Key() string
	Process(panel *ui.Panel, elm document.DocElement, value rules.PropertyValue) (string, error)
}

var FunctionMap = map[string]Function{
	"attr":                      Attr{},
	"calc":                      Calc{},
	"conic-gradient":            ConicGradient{},
	"counter":                   Counter{},
	"cubic-bezier":              CubicBezier{},
	"hsl":                       Hsl{},
	"hsla":                      Hsla{},
	"linear-gradient":           LinearGradient{},
	"max":                       Max{},
	"min":                       Min{},
	"radial-gradient":           RadialGradient{},
	"repeating-conic-gradient":  RepeatingConicGradient{},
	"repeating-linear-gradient": RepeatingLinearGradient{},
	"repeating-radial-gradient": RepeatingRadialGradient{},
	"rgb":                       Rgb{},
	"rgba":                      Rgba{},
	"var":                       Var{},
}
