/******************************************************************************/
/* css_border_left.go                                                         */
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

func (BorderLeft) Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	return mergeFutureBorderSideWidths(values, rules, borderSideLeft, "border-left")
}

// border-width border-style border-color|initial|inherit
func (BorderLeft) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 || len(values) > 3 {
		return errors.New("Border requires 1-3 values")
	}
	BorderLeftWidth{}.Process(panel, elm, values[:1], host)
	if len(values) > 1 {
		BorderLeftStyle{}.Process(panel, elm, values[1:2], host)
	}
	if len(values) > 2 {
		BorderLeftColor{}.Process(panel, elm, values[2:], host)
	}
	return nil
}
