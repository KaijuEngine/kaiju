/******************************************************************************/
/* css_border_top_right_radius.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p BorderTopRightRadius) Preprocess(values []rules.PropertyValue, ruleList []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	return preprocLeftTopRightBottom(values, ruleList, "border-radius")
}

func (p BorderTopRightRadius) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("border-top-right-radus expects exactly 1 argument")
	}
	v := helpers.NumFromLength(values[0].Str, host.Window)
	panel.SetBorderRadiusTopRight(v)
	return nil
}
