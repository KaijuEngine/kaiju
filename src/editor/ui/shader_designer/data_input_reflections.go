package shader_designer

import (
	"fmt"
	"kaiju/editor/alert"
	"kaiju/klib"
	"kaiju/markup/document"
	"kaiju/rendering"
	"kaiju/ui"
	"reflect"
	"regexp"
	"slices"
	"sort"
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

func reflectUIStructure(obj any, path string, fallbackOptions map[string][]string) DataUISection {
	section := DataUISection{}
	v := reflect.ValueOf(obj).Elem()
	vt := v.Type()
	section.Name = vt.Name()
	for i := range v.NumField() {
		f := v.Field(i)
		kind := f.Kind()
		tag := v.Type().Field(i).Tag
		if tag.Get("ignore") == "true" {
			continue
		}
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
			isList := false
			if op, ok := tag.Lookup("options"); ok && op != "" {
				keys := reflect.ValueOf(rendering.StringVkMap[op]).MapKeys()
				field.List = make([]string, len(keys))
				for i := range keys {
					field.List[i] = keys[i].String()
				}
				isList = true
			} else {
				field.List, isList = fallbackOptions[field.Name]
			}
			sort.Strings(field.List)
			if isList {
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
					myPath := fmt.Sprintf("%s.%d", p, j)
					s := reflectUIStructure(f.Index(j).Addr().Interface(), myPath, fallbackOptions)
					field.Sections = append(field.Sections, s)
				}
			} else {
				field.Type = "struct"
				s := reflectUIStructure(f.Addr().Interface(), p, fallbackOptions)
				field.Sections = append(field.Sections, s)
			}
		}
		section.Fields = append(section.Fields, field)
	}
	return section
}

func reflectAddToSlice(obj any, e *document.Element) {
	v := reflectObjectValueFromUI(obj, e)
	v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
}

func reflectRemoveFromSlice(obj any, e *document.Element) {
	ok := <-alert.New("Delete entry?", "Are you sure you want to delete this entry? The action currently can't be undone.", "Yes", "No", e.UI.Host())
	if !ok {
		return
	}
	v := reflectObjectValueFromUI(obj, e)
	index, _ := strconv.Atoi(e.Attribute("data-index"))
	v.Set(reflect.AppendSlice(v.Slice(0, index), v.Slice(index+1, v.Len())))
}
