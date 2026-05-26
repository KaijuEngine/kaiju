/******************************************************************************/
/* css_border_right_width.go                                                  */
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

// medium|thin|thick|length|initial|inherit
func (p BorderRightWidth) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	return preprocLeftTopRightBottom(values, ruleList, "border-width")
}

func (p BorderRightWidth) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("BorderRightWidth requires exactly 1 value")
	} else {
		current := panel.Base().Layout().Border()
		size := borderSizeFromStr(values[0].Str, host.Window, current.Z())
		panel.SetBorderSize(current.X(), current.Y(), size, current.W())
		return nil
	}
}
