/******************************************************************************/
/* css_background.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"errors"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Background) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return errors.New("background requires at least 1 value")
	}

	// Support:
	// 1) Single color token:
	//    background: #314052;
	//    background: rgb(49, 64, 82);
	// 2) Single image token:
	//    background: url("panel_bg.png");
	// 3) Multi-token values that include a color token:
	//    background: no-repeat center/cover #314052;
	//    background: fixed url("panel_bg.png") #202733;
	//
	// NOTE: this is intentionally partial shorthand support. Non-color/non-url
	// components are not fully expanded into individual background-* properties.
	//
	// Not supported:
	// - multiple background layers (comma-separated values)
	// - position/size parsing via slash syntax (e.g. center / cover)
	// - repeat/attachment/origin/clip token decomposition
	if len(values) == 1 {
		v := values[0]
		if strings.HasPrefix(v.Str, "url(") || (v.IsFunction() && v.Str == "url") {
			return BackgroundImage{}.Process(panel, elm, values, host)
		}
		return BackgroundColor{}.Process(panel, elm, values, host)
	}

	// CSS allows the color token to appear among other tokens.
	// Example: background: url("panel.png") no-repeat center / cover #314052;
	// Scan from right-to-left and apply the first parseable color.
	for i := len(values) - 1; i >= 0; i-- {
		if err := (BackgroundColor{}).Process(panel, elm, []rules.PropertyValue{values[i]}, host); err == nil {
			return nil
		}
	}

	return errors.New("background shorthand is only partially supported; expected a color and/or url(...)")
}
