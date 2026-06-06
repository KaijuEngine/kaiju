/******************************************************************************/
/* css_image_rendering.go                                                     */
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
	"kaijuengine.com/rendering"
)

func (p ImageRendering) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return errors.New("ImageRendering requires a value")
	}

	var err error = nil

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
