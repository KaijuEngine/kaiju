/******************************************************************************/
/* css_bottom.go                                                              */
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
func (p Bottom) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("bottom expects 1 value")
	}

	offset := panel.Base().Layout().InnerOffset().Bottom()
	parent := elm.Parent.Value()
	height := float32(host.Window.Height())
	if parent != nil {
		height = parent.UI.Layout().PixelSize().Y()
	}

	s := values[0].Str
	layout := elm.UI.Layout()
	switch s {
	case "auto":
	case "initial":
	case "inherit":
		if elm.Parent.Value() != nil {
			offset = elm.Parent.Value().UI.Layout().Offset().Y()
		}
	default:
		val := helpers.NumFromLength(values[0].Str, host.Window)
		if strings.HasSuffix(values[0].Str, "%") {
			l := panel.Base().Layout()
			if l.Ui().Entity().IsRoot() {
				return nil
			}
			pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
			l.SetInnerOffsetBottom(pLayout.PixelSize().Y() * val)
		} else {
			offset = val
		}
	}

	selfHeight := layout.PixelSize().Y()
	layout.SetInnerOffsetTop(-height + selfHeight + offset)

	return nil
}
