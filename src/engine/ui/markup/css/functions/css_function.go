/******************************************************************************/
/* css_function.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package functions

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

type Function interface {
	Key() string
	Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error)
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
