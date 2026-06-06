/******************************************************************************/
/* css_padding_right.go                                                       */
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
	"kaijuengine.com/matrix"
)

func (PaddingRight) Preprocess(values []rules.PropertyValue, rules []rules.Rule) ([]rules.PropertyValue, []rules.Rule) {
	return preprocLeftTopRightBottom(values, rules, "padding")
}

// length|initial|inherit
func (PaddingRight) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("PaddingRight: Expecting exactly one value")
	}

	padding, err := paddingSizeFromString(elm, values[0].Str, matrix.Vz, host.Window)
	if err != nil {
		return err
	}

	elm.UI.Layout().SetPadding(padding.X(), padding.Y(), padding.Z(), padding.W())
	return nil
}
