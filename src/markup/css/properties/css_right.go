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

// auto|length|initial|inherit
func (p Right) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("right expects 1 value")
	} else {
		offset := panel.Layout().InnerOffset()
		s := values[0].Str
		layout := elm.UI.Layout()
		switch s {
		case "auto":
			return errors.New("right Not implemented [auto]")
		case "initial":
			return errors.New("right Not implemented [initial]")
		case "inherit":
			if elm.HTML.Parent != nil {
				offset.SetRight(elm.HTML.Parent.DocumentElement.UI.Layout().Offset().X())
			}
		default:
			val := helpers.NumFromLength(values[0].Str, host.Window)
			if strings.HasSuffix(values[0].Str, "%") {
				panel.Layout().AddFunction(func(l *ui.Layout) {
					if elm.HTML.Parent == nil || elm.HTML.Parent.DocumentElement.UI == nil {
						return
					}
					pLayout := elm.HTML.Parent.DocumentElement.UI.Layout()
					layout.SetInnerOffsetRight(pLayout.PixelSize().X() * val)
				})
			} else {
				offset[matrix.Vz] += val
			}
		}
		layout.SetInnerOffset(offset.X(), offset.Y(), offset.Z(), offset.W())
		layout.AnchorTo(layout.Anchor().ConvertToRight())
	}
	return nil
}
