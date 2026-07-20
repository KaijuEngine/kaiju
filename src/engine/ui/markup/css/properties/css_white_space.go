/******************************************************************************/
/* css_white_space.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func setChildTextWrap(elm *document.Element, wrap bool) {
	for _, c := range elm.Children {
		if c.IsText() {
			c.UI.ToLabel().SetWrap(wrap)
		}
		setChildTextWrap(c, wrap)
	}
}

func (p WhiteSpace) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	val := values[0].Str
	wrap := true
	if val == "nowrap" {
		wrap = false
	} else if val != "normal" {
		return nil
	}

	setChildTextWrap(elm, wrap)

	return nil
}
