package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/helpers"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
	"strings"
)

func (p Width) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	var width float32
	var err error = nil
	if len(values) != 1 {
		err = fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	} else {
		width = helpers.NumFromLength(values[0].Str, host.Window)
	}
	if err == nil {
		if strings.HasSuffix(values[0].Str, "%") && elm.HTML.Parent != nil {
			panel.Layout().AddFunction(func(l *ui.Layout) {
				if elm.HTML.Parent == nil || elm.HTML.Parent.DocumentElement.UI == nil {
					return
				}
				pLayout := elm.HTML.Parent.DocumentElement.UI.Layout()
				s := pLayout.PixelSize().X()
				pPad := pLayout.Padding()
				s -= pPad.X() + pPad.Z()
				// Subtracting local padding because it's added in final scale
				p := l.Padding()
				w := s*width - p.X() - p.Z()
				l.ScaleWidth(w)
			})
		} else {
			panel.Layout().ScaleWidth(width)
		}
		panel.DontFitContentWidth()
	}
	return err
}
