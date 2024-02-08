package properties

import (
	"errors"
	"kaiju/engine"
	"kaiju/markup/css/rules"
	"kaiju/markup/document"
	"kaiju/matrix"
	"kaiju/ui"
)

// length|initial|inherit
func (p PaddingRight) Process(panel *ui.Panel, elm document.DocElement, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return errors.New("PaddingRight: Expecting exactly one value")
	}
	if padding, err := paddingSizeFromString(elm, values[0].Str, matrix.Vz, host.Window); err != nil {
		return err
	} else {
		elm.UI.Layout().SetPadding(padding.X(), padding.Y(), padding.Z(), padding.W())
		return nil
	}
}
