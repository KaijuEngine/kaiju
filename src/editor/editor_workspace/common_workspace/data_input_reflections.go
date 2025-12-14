/******************************************************************************/
/* data_input_reflections.go                                                  */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package common_workspace

import (
	"fmt"
	"kaiju/editor/codegen/reflect_helpers"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup/document"
	"kaiju/rendering"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type DataUISection struct {
	Name   string
	Fields []DataUISectionField
}

type DataUISectionField struct {
	Name     string
	Type     string
	List     []ui.SelectOption
	Value    any
	Sections []DataUISection
	RootPath string
	TipKey   string
	Label    string
}

func (f DataUISectionField) DisplayName() string {
	if f.Label != "" {
		return f.Label
	}
	return f.PascalToTitle(f.Name)
}

func (f DataUISectionField) FullPath() string {
	if f.RootPath != "" {
		return f.RootPath + "." + f.Name
	}
	return f.Name
}

func (f DataUISectionField) PascalToTitle(str string) string {
	re := regexp.MustCompile("([A-Z])")
	result := re.ReplaceAllString(str, " $1")
	return strings.TrimSpace(result)
}

func (f DataUISectionField) ValueListHas(val string) bool {
	return slices.Contains(f.Value.([]string), val)
}

func ReflectObjectValueFromUI(obj any, e *document.Element) reflect.Value {
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

func SetObjectValueFromUI(obj any, e *document.Element) {
	v := ReflectObjectValueFromUI(obj, e)
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
			res := reflect_helpers.StringToTypeValue(v.Type().String(), e.UI.ToInput().Text())
			val = reflect.ValueOf(res)
		case ui.ElementTypeSelect:
			val = reflect.ValueOf(e.UI.ToSelect().Value())
		case ui.ElementTypeCheckbox:
			val = reflect.ValueOf(e.UI.ToCheckbox().IsChecked())
		}
		v.Set(val)
	}
}

func ReflectUIStructure(obj any, path string, fallbackOptions map[string][]ui.SelectOption) DataUISection {
	section := DataUISection{}
	v := reflect.ValueOf(obj).Elem()
	vt := v.Type()
	section.Name = vt.Name()
	for i := range v.NumField() {
		f := v.Field(i)
		kind := f.Kind()
		tag := v.Type().Field(i).Tag
		if tag.Get("visible") == "false" {
			continue
		}
		field := DataUISectionField{
			Name:     vt.Field(i).Name,
			Type:     f.Type().Name(),
			Value:    f.Interface(),
			RootPath: path,
			TipKey:   tag.Get("tip"),
			Label:    tag.Get("label"),
		}
		if d := tag.Get("default"); d != "" {
			v := reflect.New(reflect.TypeOf(field.Value))
			if v.Elem().Equal(reflect.ValueOf(field.Value)) {
				field.Value = d
			}
		}
		if field.TipKey == "" {
			field.TipKey = field.Name
		}
		if (kind == reflect.String) ||
			(kind == reflect.Slice && f.Type().Elem().Kind() == reflect.String) {
			isList := false
			if op, ok := tag.Lookup("options"); ok && op != "" {
				keys := reflect.ValueOf(rendering.StringVkMap[op]).MapKeys()
				field.List = make([]ui.SelectOption, len(keys))
				for i := range keys {
					field.List[i] = ui.SelectOption{
						Name:  keys[i].String(),
						Value: keys[i].String(),
					}
				}
				isList = true
			} else {
				field.List, isList = fallbackOptions[field.Name]
			}
			sort.Slice(field.List, func(i, j int) bool {
				return field.List[i].Name < field.List[i].Name
			})
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
					s := ReflectUIStructure(f.Index(j).Addr().Interface(), myPath, fallbackOptions)
					field.Sections = append(field.Sections, s)
				}
			} else {
				field.Type = "struct"
				s := ReflectUIStructure(f.Addr().Interface(), p, fallbackOptions)
				field.Sections = append(field.Sections, s)
			}
		}
		section.Fields = append(section.Fields, field)
	}
	return section
}

func ReflectAddToSlice(obj any, e *document.Element) {
	v := ReflectObjectValueFromUI(obj, e)
	v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
}

func ReflectRemoveFromSlice(obj any, e *document.Element) {
	v := ReflectObjectValueFromUI(obj, e)
	index, _ := strconv.Atoi(e.Attribute("data-index"))
	v.Set(reflect.AppendSlice(v.Slice(0, index), v.Slice(index+1, v.Len())))
}
