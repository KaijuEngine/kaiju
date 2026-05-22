/******************************************************************************/
/* css_property_helpers.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/document"
)

func childLabels(elm *document.Element) []*ui.Label {
	labels := make([]*ui.Label, 0)
	for _, c := range elm.Children {
		if c.IsText() {
			labels = append(labels, c.UI.ToLabel())
		} else {
			labels = append(labels, childLabels(c)...)
		}
	}
	return labels
}
