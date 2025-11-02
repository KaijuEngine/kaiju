/******************************************************************************/
/* entity_data_binding.go                                                     */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package entity_data_binding

import (
	"kaiju/editor/codegen"
	"kaiju/engine"
	"kaiju/klib"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
)

var (
	tagParsers = map[string]func(f *EntityDataField, value string){
		"default": tagDefault,
		"clamp":   tagClamp,
	}
)

type EntityDataEntry struct {
	Gen       codegen.GeneratedType
	BoundData any
	Name      string
	Fields    []EntityDataField
}

type EntityDataField struct {
	Idx   int
	Name  string
	Type  string
	Pkg   string
	Value any
	Min   any
	Max   any
}

func (f *EntityDataField) IsNumber() bool   { return isNumber(f.Type) }
func (f *EntityDataField) IsInput() bool    { return isInput(f.Type) }
func (f *EntityDataField) IsCheckbox() bool { return isCheckbox(f.Type) }
func (f *EntityDataField) IsEntityId() bool { return isEntityId(f.Pkg, f.Type) }

func (de *EntityDataEntry) SetFieldByName(name string, value any) {
	f := reflect.ValueOf(de.BoundData).Elem().FieldByName(name)
	engine.ReflectEntityDataBindingValueFromJson(value, f)
}

func (de *EntityDataEntry) ReadEntityDataBindingType(g codegen.GeneratedType) *EntityDataEntry {
	v := g.New().Value
	de.Name = g.Name
	de.Gen = g
	for i := range g.Fields {
		if g.Fields[i].IsExported() {
			ef := EntityDataField{
				Idx:   i,
				Name:  g.Fields[i].Name,
				Type:  g.Fields[i].Type.Name(),
				Pkg:   g.Fields[i].Type.PkgPath(),
				Value: v.Elem().Field(i).Interface(),
			}
			if string(g.Fields[i].Tag) != "" {
				for k, fn := range tagParsers {
					if v, ok := g.Fields[i].Tag.Lookup(k); ok {
						fn(&ef, v)
						break
					}
				}
			}
			de.Fields = append(de.Fields, ef)
			fv := v.Elem().Field(len(de.Fields) - 1)
			fv.Set(reflect.ValueOf(ef.Value))
		}
	}
	de.BoundData = v.Interface()
	return de
}

func (g *EntityDataEntry) FieldNumberAsString(fieldIdx int) string {
	f := g.Fields[fieldIdx]
	if !f.IsNumber() {
		return "0"
	}
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	switch f.Value.(type) {
	case int:
		return strconv.FormatInt(v.Int(), 10)
	case int8:
		return strconv.FormatInt(v.Int(), 10)
	case int16:
		return strconv.FormatInt(v.Int(), 10)
	case int32:
		return strconv.FormatInt(v.Int(), 10)
	case int64:
		return strconv.FormatInt(v.Int(), 10)
	case uint:
		return strconv.FormatUint(v.Uint(), 10)
	case uint8:
		return strconv.FormatUint(v.Uint(), 10)
	case uint16:
		return strconv.FormatUint(v.Uint(), 10)
	case uint32:
		return strconv.FormatUint(v.Uint(), 10)
	case uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case float32:
		return klib.StripFloatStringZeros(strconv.FormatFloat(v.Float(), 'f', 5, 32))
	case float64:
		return klib.StripFloatStringZeros(strconv.FormatFloat(v.Float(), 'f', 5, 64))
	}
	return "0"
}

func (g *EntityDataEntry) FieldString(fieldIdx int) string {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return v.String()
}

func (g *EntityDataEntry) FieldBool(fieldIdx int) bool {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return v.Bool()
}

func (g *EntityDataEntry) FieldValue(fieldIdx int) any {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return v.Interface()
}

func tagDefault(f *EntityDataField, value string) {
	f.Value = klib.StringToTypeValue(f.Type, value)
}

func tagClamp(f *EntityDataField, value string) {
	if !f.IsNumber() {
		slog.Warn("cannot use the clamp tag on non-numeric field", "field", f.Name)
		return
	}
	parts := strings.Split(value, ",")
	if len(parts) == 2 {
		parts = append([]string{"0"}, parts...)
	}
	if len(parts) == 3 {
		values := make([]any, len(parts))
		for i := range parts {
			values[i] = klib.StringToTypeValue(f.Type, parts[i])
		}
		f.Value = values[0]
		f.Min = values[1]
		f.Max = values[2]
	} else {
		slog.Warn("invalid format for the 'clamp' tag on field", "field", f.Name)
	}
}

func isNumber(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128":
		return true
	default:
		return false
	}
}

func isInput(typeName string) bool {
	return typeName == "string" || isNumber(typeName)
}

func isCheckbox(typeName string) bool {
	return typeName == "bool"
}

func isEntityId(packageName, typeName string) bool {
	return packageName == "kaiju/engine" && typeName == "EntityId"
}
