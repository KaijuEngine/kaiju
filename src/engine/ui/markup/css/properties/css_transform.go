/******************************************************************************/
/* css_transform.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/helpers"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"strings"
)

func translateXYZ(str string, panel *ui.Panel, host *engine.Host, vc matrix.VectorComponent) {
	if vc == matrix.Vz {
		p := panel.Base().Entity().Transform.Position()
		p[vc] += helpers.NumFromLength(str, host.Window)
		panel.Base().Entity().Transform.SetPosition(p)
	} else {
		offset := panel.Base().Layout().InnerOffset()
		p := helpers.NumFromLength(str, host.Window)
		if vc == matrix.Vy {
			p *= -1.0
		}
		if strings.HasSuffix(str, "%") {
			l := panel.Base().Layout()
			localInnerOffset := l.LocalInnerOffset()
			localInnerOffset[vc] = l.PixelSize()[vc] * p
			l.SetLocalInnerOffset(localInnerOffset.X(), localInnerOffset.Y(), localInnerOffset.Z(), localInnerOffset.W())
		} else {
			offset[vc] += p
			panel.Base().Layout().SetInnerOffset(offset.X(), offset.Y(), offset.Z(), offset.W())
		}
	}
}

// none|transform-functions|initial|inherit
func (p Transform) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("transform expects 1 value")
	}
	switch values[0].Str {
	case "none":
	case "initial":
	case "inherit":
	case "matrix":
	case "matrix3d":
	case "translate":
		if len(values[0].Args) == 2 {
			translateXYZ(values[0].Args[0], panel, host, matrix.Vx)
			translateXYZ(values[0].Args[1], panel, host, matrix.Vy)
		} else {
			return errors.New("translate expects 2 values")
		}
	case "translate3d":
		if len(values[0].Args) == 3 {
			translateXYZ(values[0].Args[0], panel, host, matrix.Vx)
			translateXYZ(values[0].Args[1], panel, host, matrix.Vy)
			translateXYZ(values[0].Args[2], panel, host, matrix.Vz)
		} else {
			return errors.New("translate3d expects 3 values")
		}
	case "translateX":
		translateXYZ(values[0].Args[0], panel, host, matrix.Vx)
	case "translateY":
		translateXYZ(values[0].Args[0], panel, host, matrix.Vy)
	case "translateZ":
		translateXYZ(values[0].Args[0], panel, host, matrix.Vz)
	case "scale":
	case "scale3d":
	case "scaleX":
	case "scaleY":
	case "scaleZ":
	case "rotate":
	case "rotate3d":
	case "rotateX":
	case "rotateY":
	case "rotateZ":
	case "skew":
	case "skewX":
	case "skewY":
	case "perspective":
	default:
		return errors.New("transform has unexpected value")
	}
	return nil
}
