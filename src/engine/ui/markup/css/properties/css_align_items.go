/******************************************************************************/
/* css_align_items.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"fmt"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p AlignItems) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) == 0 {
		return nil
	}
	align, ok := parseFlexAlign(values[0].Str)
	if !ok || align == ui.FlexAlignAuto {
		return fmt.Errorf("invalid align-items value %q", values[0].Str)
	}
	panel.SetFlexAlignItems(align)
	return nil
}
