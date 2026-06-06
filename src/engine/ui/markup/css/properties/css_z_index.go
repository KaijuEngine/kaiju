/******************************************************************************/
/* css_z_index.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"strconv"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p ZIndex) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("ZIndex requires exactly 1 value")
	} else {
		val, _ := strconv.ParseFloat(values[0].Str, 64)
		z := float32(val)
		p := elm.Parent.Value()
		for p != nil && p.UI != nil {
			z += p.UI.Layout().Z()
			p = p.Parent.Value()
		}
		panel.Base().Layout().SetZ(z)
		return nil
	}
}
