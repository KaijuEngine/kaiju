/******************************************************************************/
/* css_kaiju_shader.go                                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"
	"regexp"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p KaijuMaterial) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("Expected exactly 1 value but got %d", len(values))
	}

	reg := regexp.MustCompile(`url\s{0,}\(\s{0,}"(.*?)"\s{0,}\)`)
	parts := reg.FindStringSubmatch(values[0].Str)
	if len(parts) != 2 {
		return fmt.Errorf("Expected exactly 1 url but got %d", len(parts)-1)
	}

	path := strings.TrimSpace(parts[1])
	mat, err := host.MaterialCache().Material(path)
	if err != nil {
		return err
	}

	panel.SetMaterial(mat)
	return nil
}
