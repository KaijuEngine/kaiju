/******************************************************************************/
/* css_margin_left.go                                                         */
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

func (MarginLeft) Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	return preprocLeftTopRightBottom(values, rules, "margin")
}

// length|auto|initial|inherit
func (MarginLeft) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("MarginLeft requires exactly 1 value")
	}

	current := panel.Base().Layout().Margin()
	size := marginSizeFromStr(values[0].Str, host.Window)
	panel.Base().Layout().SetMargin(size, current.Y(), current.Z(), current.W())
	return nil
}
