package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/ui"
	"kaiju/uimarkup/css/helpers"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
	"strings"
)

// auto|length|initial|inherit
func (p Left) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("left expects 1 value")
	} else {
		offset := panel.Layout().InnerOffset()
		s := values[0].Str
		layout := elm.UI.Layout()
		switch s {
		case "auto":
			return errors.New("Left Not implemented [auto]")
		case "initial":
			return errors.New("Left Not implemented [initial]")
		case "inherit":
			if elm.HTML.Parent != nil {
				offset[matrix.Vx] += elm.HTML.Parent.DocumentElement.UI.Layout().Offset().X()
			}
		default:
			val := helpers.NumFromLength(values[0].Str, host.Window)
			if strings.HasSuffix(values[0].Str, "%") {
				panel.Layout().AddFunction(func(l *ui.Layout) {
					if elm.HTML.Parent == nil || elm.HTML.Parent.DocumentElement.UI == nil {
						return
					}
					pLayout := elm.HTML.Parent.DocumentElement.UI.Layout()
					layout.SetInnerOffsetLeft(pLayout.PixelSize().X() * val)
				})
			} else {
				offset[matrix.Vx] += val
			}
		}
		layout.SetInnerOffset(offset.X(), offset.Y(), offset.Z(), offset.W())
		layout.AnchorTo(layout.Anchor().ConvertToLeft())
	}
	return nil
}
