package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
)

func (p ImageRendering) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) (err error) {
	if len(values) == 0 {
		return errors.New("ImageRendering requires a value")
	}
	switch values[0].Str {
	case "crisp-edges":
		tex := panel.Background()
		if tex.Filter != rendering.TextureFilterLinear {
			tex, err = host.TextureCache().Texture(tex.Key, rendering.TextureFilterLinear)
			if err == nil {
				panel.SetBackground(tex)
			}
		}
	case "pixelated":
		tex := panel.Background()
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
