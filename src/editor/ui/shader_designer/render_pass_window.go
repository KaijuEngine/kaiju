package shader_designer

import (
	"kaiju/klib"
	"kaiju/markup/document"
	"kaiju/ui"
	"reflect"
	"strconv"
	"strings"
)

func (win *ShaderDesigner) renderPassValueChanged(e *document.Element) {
	id := e.Attribute("id")
	idx := -1
	sep := strings.Index(id, "_")
	if sep >= 0 {
		if i, err := strconv.Atoi(id[sep+1:]); err == nil {
			idx = i
		}
		id = id[:sep]
	}
	var v reflect.Value
	if idx >= 0 {
		v = reflect.ValueOf(&win.pipeline.ColorBlendAttachments[idx])
	} else {
		v = reflect.ValueOf(&win.pipeline)
	}
	field := v.Elem().FieldByName(id)
	var val reflect.Value
	switch e.UI.Type() {
	case ui.ElementTypeInput:
		res := klib.StringToTypeValue(field.Type().String(), e.UI.ToInput().Text())
		val = reflect.ValueOf(res)
	case ui.ElementTypeSelect:
		val = reflect.ValueOf(e.UI.ToSelect().Value())
	case ui.ElementTypeCheckbox:
		val = reflect.ValueOf(e.UI.ToCheckbox().IsChecked())
	}
	field.Set(val)
}
