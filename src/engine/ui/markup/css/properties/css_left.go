/******************************************************************************/
/* css_left.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

// auto|length|initial|inherit
func (p Left) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("left expects 1 value")
	}
	offsetX := panel.Base().Layout().InnerOffset().Left()
	s := values[0].Str
	layout := elm.UI.Layout()
	switch s {
	case "auto":
		return errors.New("Left Not implemented [auto]")
	case "initial":
		return errors.New("Left Not implemented [initial]")
	case "inherit":
		if elm.Parent.Value() != nil {
			offsetX += elm.Parent.Value().UI.Layout().Offset().X()
		}
	default:
		val := values[0].Num
		if strings.HasSuffix(values[0].Str, "%") {
			l := panel.Base().Layout()
			if l.Ui().Entity().IsRoot() {
				return nil
			}
			pLayout := ui.FirstOnEntity(l.Ui().Entity().Parent).Layout()
			offsetX = pLayout.ContentSize().X() * val
		} else {
			offsetX = val
		}
	}
	layout.SetInnerOffsetLeft(offsetX)
	return nil
}
