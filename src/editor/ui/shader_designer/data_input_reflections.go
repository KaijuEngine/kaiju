package shader_designer

import (
	"kaiju/klib"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	dataInputHTML = "editor/ui/shader_designer/data_input_window.html"
)

type DataUISection struct {
	Name   string
	Fields []DataUISectionField
}

type DataUISectionField struct {
	Name     string
	Type     string
	List     []string
	Value    any
	Sections []DataUISection
	RootPath string
	TipKey   string
}

func (f DataUISectionField) DisplayName() string {
	re := regexp.MustCompile("([A-Z])")
	result := re.ReplaceAllString(f.Name, " $1")
	return strings.TrimSpace(result)
}

func (f DataUISectionField) FullPath() string {
	if f.RootPath != "" {
		return f.RootPath + "." + f.Name
	}
	return f.Name
}

func (f DataUISectionField) ValueListHas(val string) bool {
	return slices.Contains(f.Value.([]string), val)
}

func reflectObjectValueFromUI(obj any, e *document.Element) reflect.Value {
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
	return v
}

func setObjectValueFromUI(obj any, e *document.Element) {
	v := reflectObjectValueFromUI(obj, e)
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

func reflectUIStructure(obj any, path string) DataUISection {
	section := DataUISection{}
	v := reflect.ValueOf(obj).Elem()
	vt := v.Type()
	section.Name = vt.Name()
	for i := range v.NumField() {
		f := v.Field(i)
		kind := f.Kind()
		tag := v.Type().Field(i).Tag
		field := DataUISectionField{
			Name:     vt.Field(i).Name,
			Type:     f.Type().Name(),
			Value:    f.Interface(),
			RootPath: path,
			TipKey:   tag.Get("tip"),
		}
		if field.TipKey == "" {
			field.TipKey = field.Name
		}
		if (kind == reflect.String) ||
			(kind == reflect.Slice && f.Type().Elem().Kind() == reflect.String) {
			if op, ok := tag.Lookup("options"); ok {
				keys := reflect.ValueOf(rendering.StringVkMap[op]).MapKeys()
				field.List = make([]string, len(keys))
				for i := range keys {
					field.List[i] = keys[i].String()
				}
				if kind == reflect.String {
					field.Type = "enum"
				} else {
					field.Type = "bitmask"
				}
			}
		} else if kind == reflect.Slice || kind == reflect.Struct {
			p := field.FullPath()
			if kind == reflect.Slice {
				field.Type = "slice"
				childCount := f.Len()
				for j := range childCount {
					s := reflectUIStructure(f.Index(j).Addr().Interface(), p)
					field.Sections = append(field.Sections, s)
				}
			} else {
				field.Type = "struct"
				s := reflectUIStructure(f.Addr().Interface(), p)
				field.Sections = append(field.Sections, s)
			}
		}
		section.Fields = append(section.Fields, field)
	}
	return section
}
