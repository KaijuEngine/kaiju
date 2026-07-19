/******************************************************************************/
/* css_word_wrap.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func setChildTextWordWrap(elm *document.Element, wrap bool) {
	if elm == nil {
		return
	}
	if elm.UI == nil {
		for _, child := range elm.Children {
			setChildTextWordWrap(child, wrap)
		}
		return
	}
	if elm.UI.IsType(ui.ElementTypeInput) {
		elm.UI.ToInput().SetWrap(wrap)
		return
	}
	if elm.UI.IsType(ui.ElementTypeTextArea) {
		elm.UI.ToTextArea().SetWrap(wrap)
		return
	}
	for _, c := range elm.Children {
		if c.IsText() {
			c.UI.ToLabel().SetWrap(wrap)
		} else {
			setChildTextWordWrap(c, wrap)
		}
	}
}

func (p WordWrap) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("WordWrap requires a single value")
	}

	switch values[0].Str {
	case "normal":
		setChildTextWordWrap(elm, true)
	case "unset":
		setChildTextWordWrap(elm, false)
	case "inherit":
	case "initial":
	case "break-word":
		// TODO:  Implement word breaking in labels
		fallthrough
	default:
		return errors.New("WordWrap does not currently support " + values[0].Str)
	}

	return nil
}
