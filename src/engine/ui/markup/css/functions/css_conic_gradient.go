/******************************************************************************/
/* css_conic_gradient.go                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package functions

import (
	"errors"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (f ConicGradient) Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error) {
	return "", errors.New("not implemented")
}
