package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/rendering"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
	"regexp"
	"strings"
)

func (p BackgroundImage) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("Expected exactly 1 value but got " + string(len(values)))
	}
	reg := regexp.MustCompile(`url\s{0,}\(\s{0,}"(.*?)"\s{0,}\)`)
	if parts := reg.FindStringSubmatch(values[0].Str); len(parts) != 2 {
		return errors.New("Expected exactly 1 url but got " + string(len(parts)-1))
	} else {
		path := strings.TrimSpace(parts[1])
		if tex, err := host.TextureCache().Texture(path, rendering.TextureFilterLinear); err != nil {
			return err
		} else {
			panel.SetBackground(tex)
			return nil
		}
	}
}
