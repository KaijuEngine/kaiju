/******************************************************************************/
/* css_opacity.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"fmt"
	"strconv"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
)

func (p Opacity) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if !panel.HasBackground() {
		return errors.New("can't set opacity on panel without an image/color")
	}
	if len(values) != 1 {
		return fmt.Errorf("expected 1 argument to opacity but got %d", len(values))
	}
	opacity, err := strconv.ParseFloat(values[0].Str, 64)
	if err != nil {
		return err
	}
	c := panel.Color()
	c.SetA(klib.Clamp(matrix.Float(opacity), 0, 1))
	panel.SetColor(c)
	return nil
}
