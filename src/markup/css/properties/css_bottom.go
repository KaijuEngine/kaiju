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
func (p Bottom) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("bottom expects 1 value")
	} else {
		offset := panel.Layout().InnerOffset()
		s := values[0].Str
		layout := elm.UI.Layout()
		switch s {
		case "auto":
		case "initial":
		case "inherit":
			if elm.HTML.Parent != nil {
				offset.SetBottom(elm.HTML.Parent.DocumentElement.UI.Layout().Offset().Y())
			}
		default:
			val := helpers.NumFromLength(values[0].Str, host.Window)
			if strings.HasSuffix(values[0].Str, "%") {
				panel.Layout().AddFunction(func(l *ui.Layout) {
					if elm.HTML.Parent == nil || elm.HTML.Parent.DocumentElement.UI == nil {
						return
					}
					pLayout := elm.HTML.Parent.DocumentElement.UI.Layout()
					layout.SetInnerOffsetBottom(pLayout.PixelSize().Y() * val)
				})
			} else {
				offset[matrix.Vw] += val
			}
		}
		layout.SetInnerOffset(offset.Left(), offset.Top(), offset.Right(), -offset.Bottom())
		layout.AnchorTo(layout.Anchor().ConvertToBottom())
	}
	return nil
}
