/******************************************************************************/
/* css_property_helpers.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package properties

import (
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/rules"
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

func expandFourSideValues(values []rules.PropertyValue) []rules.PropertyValue {
	values = clonePropertyValues(values)
	switch len(values) {
	case 1:
		for i := range 3 {
			values = append(values, values[i].Clone())
		}
	case 2:
		values = append(values, values[0].Clone())
		values = append(values, values[1].Clone())
	case 3:
		values = append(values, values[1].Clone())
	}
	return values
}

func clonePropertyValues(values []rules.PropertyValue) []rules.PropertyValue {
	out := make([]rules.PropertyValue, len(values))
	for i := range values {
		out[i] = values[i].Clone()
	}
	return out
}
