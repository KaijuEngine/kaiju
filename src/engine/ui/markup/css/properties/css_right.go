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
	"kaijuengine.com/engine/ui/markup/css/functions"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

// auto|length|initial|inherit
func (p Right) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("right expects 1 value")
	}

	offset := panel.Base().Layout().InnerOffset().Right()
	parent := elm.Parent.Value()
	width := matrix.Float(host.Window.Width())
	if parent != nil {
		parentLayout := parent.UI.Layout()
		width = parentLayout.PixelSize().X() - parentLayout.Border().Horizontal()
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
			offset = pLayout.PixelSize().X() * val
		} else if values[0].IsFunction() {
			if values[0].Str == "calc" {
				val := values[0]
				val.Args = append(val.Args, "width")
				res, _ := functions.Calc{}.Process(panel, elm, val)
				offset = helpers.NumFromLength(res, host.Window)
			}
		} else {
			offset = val
		}
	}
	selfWidth := layout.PixelSize().X()
	layout.SetInnerOffsetLeft(width - selfWidth - offset)

	return nil
}
