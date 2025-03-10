/******************************************************************************/
/* css_image_rendering.go                                                     */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
)

func (p ImageRendering) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	var err error = nil
	if len(values) == 0 {
		return errors.New("ImageRendering requires a value")
	}
	switch values[0].Str {
	case "crisp-edges":
		tex := panel.Background()
		if tex == nil {
			return errors.New("Failed to set image rendering, no background created yet, possibly CSS sort order issue")
		}
		if tex.Filter != rendering.TextureFilterLinear {
			tex, err = host.TextureCache().Texture(tex.Key, rendering.TextureFilterLinear)
			if err == nil {
				panel.SetBackground(tex)
			}
		}
	case "pixelated":
		tex := panel.Background()
		if tex == nil {
			return errors.New("Failed to set image rendering, no background created yet, possibly CSS sort order issue")
		}
		if tex.Filter != rendering.TextureFilterNearest {
			tex, err = host.TextureCache().Texture(tex.Key, rendering.TextureFilterNearest)
			if err == nil {
				panel.SetBackground(tex)
			}
		}
	default:
		return errors.New("invalid value for ImageRendering")
	}
	return nil
}
