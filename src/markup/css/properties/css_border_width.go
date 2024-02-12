/*****************************************************************************/
/* css_border_width.go                                                       */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* MIT License                                                               */
/*                                                                           */
/* Copyright (c) 2023-present Kaiju Engine contributors (CONTRIBUTORS.md).   */
/* Copyright (c) 2015-2023 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* Everyone who drinks of this water will be thirsty again; but whoever      */
/* drinks of the water that I will give him shall never thirst; John 4:13-14 */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining a   */
/* copy of this software and associated documentation files (the "Software"),*/
/* to deal in the Software without restriction, including without limitation */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,  */
/* and/or sell copies of the Software, and to permit persons to whom the     */
/* Software is furnished to do so, subject to the following conditions:      */
/*                                                                           */
/* The above copyright, blessing, biblical verse, notice and                 */
/* this permission notice shall be included in all copies or                 */
/* substantial portions of the Software.                                     */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS   */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY   */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE     */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                             */
/*****************************************************************************/

package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/helpers"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/ui"
)

func (p BorderWidth) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	b := [4]float32{}
	if len(values) == 1 {
		b[0] = helpers.NumFromLength(values[0].Str, host.Window)
		b[1] = b[0]
		b[2] = b[0]
		b[3] = b[0]
	} else if len(values) == 2 {
		// Top/bottom left/right
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[0] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = b[1]
	} else if len(values) == 3 {
		// Top left/right bottom
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[0] = helpers.NumFromLength(values[1].Str, host.Window)
		b[2] = b[0]
		b[3] = helpers.NumFromLength(values[2].Str, host.Window)
	} else if len(values) == 4 {
		b[1] = helpers.NumFromLength(values[0].Str, host.Window)
		b[2] = helpers.NumFromLength(values[1].Str, host.Window)
		b[3] = helpers.NumFromLength(values[2].Str, host.Window)
		b[0] = helpers.NumFromLength(values[3].Str, host.Window)
	} else {
		return errors.New("Invalid number of values for BorderRadius")
	}
	panel.SetBorderSize(b[0], b[1], b[2], b[3])
	return nil
}
