package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/matrix"
	"kaiju/ui"
	"kaiju/uimarkup/css/rules"
	"kaiju/uimarkup/markup"
)

// length|initial|inherit
func (p PaddingTop) Process(panel *ui.Panel, elm markup.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("PaddingTop: Expecting exactly one value")
	}
	if padding, err := paddingSizeFromString(elm, values[0].Str, matrix.Vy, host.Window); err != nil {
		return err
	} else {
		elm.UI.Layout().SetPadding(padding.X(), padding.Y(), padding.Z(), padding.W())
		return nil
	}
}
