package properties

import (
	"fmt"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
)

func childLabels(elm document.DocElement) []*ui.Label {
	labels := make([]*ui.Label, 0)
	for _, c := range elm.HTML.Children {
		if c.DocumentElement.HTML.IsText() {
			labels = append(labels, c.DocumentElement.UI.(*ui.Label))
		} else {
			labels = append(labels, childLabels(*c.DocumentElement)...)
		}
	}
	return labels
}

// left|right|center|justify|initial|inherit
func (p TextAlign) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}
	labels := childLabels(elm)
	switch values[0].Str {
	case "left":
		for _, l := range labels {
			l.Layout().AnchorTo(l.Layout().Anchor().ConvertToLeft())
			l.SetJustify(rendering.FontJustifyLeft)
		}
	case "right":
		for _, l := range labels {
			l.Layout().AnchorTo(l.Layout().Anchor().ConvertToRight())
			l.SetJustify(rendering.FontJustifyRight)
		}
	case "center":
		for _, l := range labels {
			l.Layout().AnchorTo(l.Layout().Anchor().ConvertToCenter())
			l.SetJustify(rendering.FontJustifyCenter)
		}
	case "justify":
		// TODO:  Support text justification
	case "initial":
	case "inherit":
	}
	return nil
}
