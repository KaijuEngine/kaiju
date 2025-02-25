package shader_designer

import (
	"kaiju/klib"
	"kaiju/markup/document"
	"kaiju/ui"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func setObjectValueFromUI(obj any, e *document.Element) {
	path := e.Attribute("data-path")
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(obj).Elem()
	for i := range parts {
		if idx, err := strconv.Atoi(parts[i]); err == nil {
			v = v.Index(idx)
		} else {
			v = v.FieldByName(parts[i])
		}
	}
	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.String {
		// TODO:  Ensure switch e.UI.Type() == ui.ElementTypeCheckbox
		add := e.UI.ToCheckbox().IsChecked()
		str := e.Attribute("name")
		var slice []string
		if !v.IsNil() {
			slice = v.Interface().([]string)
		} else {
			slice = []string{}
		}
		if add {
			for _, s := range slice {
				if s == str {
					return // Already exists, no change
				}
			}
			slice = append(slice, str)
		} else {
			for i, s := range slice {
				if s == str {
					slice = slices.Delete(slice, i, i+1)
					break
				}
			}
		}
		v.Set(reflect.ValueOf(slice))
	} else {
		var val reflect.Value
		switch e.UI.Type() {
		case ui.ElementTypeInput:
			res := klib.StringToTypeValue(v.Type().String(), e.UI.ToInput().Text())
			val = reflect.ValueOf(res)
		case ui.ElementTypeSelect:
			val = reflect.ValueOf(e.UI.ToSelect().Value())
		case ui.ElementTypeCheckbox:
			val = reflect.ValueOf(e.UI.ToCheckbox().IsChecked())
		}
		v.Set(val)
	}
}
