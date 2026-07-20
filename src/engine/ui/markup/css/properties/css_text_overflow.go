/******************************************************************************/
/* css_text_overflow.go                                                       */
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

func setChildTextOverflow(elm *document.Element, ellipsis bool) {
	for _, c := range elm.Children {
		if c.IsText() {
			c.UI.ToLabel().SetTextOverflowEllipsis(ellipsis)
		}
		setChildTextOverflow(c, ellipsis)
	}
}

func (p TextOverflow) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	val := values[0].Str
	if val == "ellipsis" {
		setChildTextOverflow(elm, true)
	} else if val == "clip" {
		setChildTextOverflow(elm, false)
	}
	
	return nil
}
