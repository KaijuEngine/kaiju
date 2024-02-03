package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/ui"
	"kaiju/uimarkup/css/functions"
	"kaiju/uimarkup/css/helpers"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
	"strings"
)

func (p Height) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	var height float32
	var err error = nil
	if len(values) != 1 {
		err = fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	} else {
		height = helpers.NumFromLength(values[0].Str, host.Window)
	}
	if err == nil {
		if strings.HasSuffix(values[0].Str, "%") {
			panel.Layout().AddFunction(func(l *ui.Layout) {
				if elm.HTML.Parent == nil || elm.HTML.Parent.DocumentElement.UI == nil {
					return
				}
				pLayout := elm.HTML.Parent.DocumentElement.UI.Layout()
				s := pLayout.PixelSize().Y()
				pPad := pLayout.Padding()
				s -= pPad.Y() + pPad.W()
				// Subtracting local padding because it's added in final scale
				p := l.Padding()
				h := s*height - p.Y() - p.W()
				l.ScaleHeight(h)
			})
		} else if values[0].IsFunction() {
			if values[0].Str == "calc" {
				panel.Layout().AddFunction(func(l *ui.Layout) {
					val := values[0]
					val.Args = append(val.Args, "height")
					res, _ := functions.Calc{}.Process(panel, elm, val)
					height = helpers.NumFromLength(res, host.Window)
					panel.Layout().ScaleHeight(height)
				})
			}
		} else {
			panel.Layout().ScaleHeight(height)
		}
		panel.DontFitContent()
	}
	return err
}
