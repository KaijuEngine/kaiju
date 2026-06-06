/******************************************************************************/
/* css_top.go                                                                 */
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
)

// auto|length|initial|inherit
func (p Top) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("top expects 1 value")
	}

	yOffset := panel.Base().Layout().InnerOffset().Top()
	layout := elm.UI.Layout()

	s := values[0].Str
	switch s {
	case "auto":
	case "initial":
	case "inherit":
		if elm.Parent.Value() != nil {
			yOffset = elm.Parent.Value().UI.Layout().Offset().Y()
		}
	default:
		val := helpers.NumFromLength(values[0].Str, host.Window)
		l := panel.Base().Layout()
		if strings.HasSuffix(values[0].Str, "%") {
			if l.Ui().Entity().IsRoot() {
				return nil
			}
			pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
			yOffset = pLayout.PixelSize().Y() * -val
		} else if values[0].IsFunction() {
			if values[0].Str == "calc" {
				val := values[0]
				val.Args = append(val.Args, "height")
				res, _ := functions.Calc{}.Process(panel, elm, val)
				top := helpers.NumFromLength(res, host.Window)
				yOffset = -top
			}
		} else {
			yOffset = -val
		}
	}

	layout.SetInnerOffsetTop(yOffset)
	return nil
}
