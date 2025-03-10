package properties

import (
	"kaiju/markup/document"
	"kaiju/ui"
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
