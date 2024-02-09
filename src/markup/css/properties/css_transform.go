package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
	"strings"
)

func translateXYZ(str string, panel *ui.Panel, host *engine.Host, vc matrix.VectorComponent) {
	if vc == matrix.Vz {
		p := panel.Entity().Transform.Position()
		p[vc] += helpers.NumFromLength(str, host.Window)
		panel.Entity().Transform.SetPosition(p)
	} else {
		offset := panel.Layout().InnerOffset()
		p := helpers.NumFromLength(str, host.Window)
		if vc == matrix.Vy {
			p *= -1.0
		}
		if strings.HasSuffix(str, "%") {
			panel.Layout().AddFunction(func(l *ui.Layout) {
				localInnerOffset := l.LocalInnerOffset()
				localInnerOffset[vc] = l.PixelSize()[vc] * p
				l.SetLocalInnerOffset(localInnerOffset.X(), localInnerOffset.Y(), localInnerOffset.Z(), localInnerOffset.W())
			})
		} else {
			offset[vc] += p
			panel.Layout().SetInnerOffset(offset.X(), offset.Y(), offset.Z(), offset.W())
		}
	}
}

// none|transform-functions|initial|inherit
func (p Transform) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
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
