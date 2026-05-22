/******************************************************************************/
/* css_right.go                                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

// auto|length|initial|inherit
func (p Right) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("right expects 1 value")
	}

	offset := panel.Base().Layout().InnerOffset().Right()
	parent := elm.Parent.Value()
	pad := float32(0)
	border := float32(0)
	width := float32(host.Window.Width())
	if parent != nil {
		width = parent.UI.Layout().PixelSize().X()
		pad = parent.UI.Layout().Padding().Right()
		border = parent.UI.Layout().Border().Right()
	}

	s := values[0].Str
	layout := elm.UI.Layout()
	switch s {
	case "auto":
		return errors.New("right Not implemented [auto]")
	case "initial":
		return errors.New("right Not implemented [initial]")
	case "inherit":
		if elm.Parent.Value() != nil {
			offset = elm.Parent.Value().UI.Layout().Offset().X()
		}
	default:
		val := helpers.NumFromLength(values[0].Str, host.Window)
		if strings.HasSuffix(values[0].Str, "%") {
			l := panel.Base().Layout()
			if l.Ui().Entity().IsRoot() {
				return nil
			}
			pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
			l.SetInnerOffsetRight(pLayout.PixelSize().X() * val)
		} else {
			offset = val
		}
	}
	selfWidth := layout.PixelSize().X()
	layout.SetInnerOffsetLeft(width - selfWidth - offset - pad - border)

	return nil
}
