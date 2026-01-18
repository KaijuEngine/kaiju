package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/css/rules"
	"kaiju/engine/ui/markup/document"
	"regexp"
	"strings"
)

func (p KaijuMaterial) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	}
	reg := regexp.MustCompile(`url\s{0,}\(\s{0,}"(.*?)"\s{0,}\)`)
	if parts := reg.FindStringSubmatch(values[0].Str); len(parts) != 2 {
		return fmt.Errorf("Expected exactly 1 url but got %d", len(parts)-1)
	} else {
		path := strings.TrimSpace(parts[1])
		mat, err := host.MaterialCache().Material(path)
		if err != nil {
			return err
		}
		panel.SetMaterial(mat)
		return nil
	}
}
