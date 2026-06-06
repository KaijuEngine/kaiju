/******************************************************************************/
/* css_rgb.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package functions

import (
	"strconv"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
	"kaijuengine.com/matrix"
)

func (f Rgb) Process(panel *ui.Panel, elm *document.Element, value rules.PropertyValue) (string, error) {
	r, _ := strconv.Atoi(value.Args[0])
	g, _ := strconv.Atoi(value.Args[1])
	b, _ := strconv.Atoi(value.Args[2])
	c := matrix.NewColor8(uint8(r), uint8(g), uint8(b), 255)
	return c.Hex(), nil
}
