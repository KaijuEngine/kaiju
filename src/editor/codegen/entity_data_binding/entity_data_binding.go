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
	"kaiju/editor/codegen/reflect_helpers"
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
func (f *EntityDataField) IsVec3() bool     { return isVec3(f.Type) }
func (f *EntityDataField) IsCheckbox() bool { return isCheckbox(f.Type) }

func (f *EntityDataField) IsEntityId() bool { return isEntityId(f.Pkg, f.Type) }

// SetFieldByName sets the struct field identified by `name` to `value`.
// It reflects into the bound data instance (de.BoundData), finds the
// exported field, and uses engine.ReflectEntityDataBindingValueFromJson
// to convert the JSON‑compatible value to the field's concrete type.
// The underlying bound struct is mutated. Panics if the field does not
// exist or cannot be set.
func (de *EntityDataEntry) SetFieldByName(name string, value any) {
	f := reflect.ValueOf(de.BoundData).Elem().FieldByName(name)
	engine.ReflectEntityDataBindingValueFromJson(value, f)
}

// ReadEntityDataBindingType populates the EntityDataEntry with information
// from the provided GeneratedType. It sets the entry's name, generated
// type reference, and creates a slice of EntityDataField values for each
// exported struct field. Tag values are parsed (e.g., "default", "clamp")
// and applied to the corresponding field. The underlying bound data
// instance is created and stored in de.BoundData, and each field's value
// is written back to the new instance. The method returns the updated
// *EntityDataEntry.
//
// Parameters:
//
//	g - a codegen.GeneratedType describing the struct to bind.
//
// Side effects:
//   - de.Name, de.Gen, de.Fields, and de.BoundData are modified.
//   - The newly created struct instance is populated with default or
//     clamped values based on struct tags.
//
// Returns:
//
//	The same *EntityDataEntry pointer for chaining.
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

// FieldNumberAsString returns the numeric value of the field at the given
// index as a string. If the field is not a numeric type, "0" is returned.
// The function reflects into the bound data struct, extracts the value,
// and formats it according to its concrete type (int, uint, float, etc.).
// For floating‑point numbers the string is trimmed of trailing zeros.
// No side effects; the bound data is only read.
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

func (g *EntityDataEntry) FieldNumber(fieldIdx int) float64 {
	f := g.Fields[fieldIdx]
	if !f.IsNumber() {
		return 0
	}
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	switch f.Value.(type) {
	case int, int8, int16, int32, int64:
		return float64(v.Int())
	case uint, uint8, uint16, uint32, uint64:
		return float64(v.Uint())
	case float32, float64:
		return v.Float()
	default:
		return 0
	}
}

func (g *EntityDataEntry) FieldInteger(fieldIdx int) int64 {
	f := g.Fields[fieldIdx]
	if !f.IsNumber() {
		return 0
	}
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	switch f.Value.(type) {
	case int, int8, int16, int32, int64:
		return v.Int()
	case uint, uint8, uint16, uint32, uint64:
		return int64(v.Uint())
	case float32, float64:
		return int64(v.Float())
	default:
		return 0
	}
}

func (g *EntityDataEntry) FieldUnsignedInteger(fieldIdx int) uint64 {
	f := g.Fields[fieldIdx]
	if !f.IsNumber() {
		return 0
	}
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	switch f.Value.(type) {
	case int, int8, int16, int32, int64:
		return uint64(v.Int())
	case uint, uint8, uint16, uint32, uint64:
		return v.Uint()
	case float32, float64:
		return uint64(v.Float())
	default:
		return 0
	}
}

// FieldString returns the string representation of the field at the given
// index. It reflects into the bound data struct, retrieves the field value,
// and calls its String method (or the underlying reflect.String conversion).
// Parameters:
//
//	fieldIdx - zero‑based index of the field within the EntityDataEntry.
//
// Returns:
//
//	The field's value as a string. If the field's type does not implement
//	String, the zero‑value string is returned.
//
// Side effects:
//
//	The function only reads data; it does not modify the bound struct.
func (g *EntityDataEntry) FieldString(fieldIdx int) string {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return v.String()
}

func (g *EntityDataEntry) FieldVectorComponent(fieldIdx, componentIdx int) float32 {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return float32(v.Index(componentIdx).Float())
}

func (g *EntityDataEntry) FieldVectorComponentAsString(fieldIdx, componentIdx int) string {
	v := g.FieldVectorComponent(fieldIdx, componentIdx)
	return klib.StripFloatStringZeros(strconv.FormatFloat(float64(v), 'f', 5, 32))
}

// FieldBool returns the boolean value of the field at the given index.
// It reflects into the bound data struct, retrieves the field, and
// returns its bool representation. No side‑effects; the bound data is
// only read.
func (g *EntityDataEntry) FieldBool(fieldIdx int) bool {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return v.Bool()
}

// FieldValue returns the underlying value of the field at the given index.
// It reflects into the bound data struct and returns the field as an
// interface{}. No modifications are made to the struct.
func (g *EntityDataEntry) FieldValue(fieldIdx int) any {
	v := reflect.ValueOf(g.BoundData).Elem().Field(fieldIdx)
	return v.Interface()
}

// FieldValueByName returns the value of the struct field identified by
// name. It reflects into the bound data instance, looks up the field by
// its name, and returns the field's value as an interface{}. The bound
// data is read‑only; no modifications are performed.
func (g *EntityDataEntry) FieldValueByName(name string) any {
	v := reflect.ValueOf(g.BoundData).Elem().FieldByName(name)
	return v.Interface()
}

func tagDefault(f *EntityDataField, value string) {
	f.Value = reflect_helpers.StringToTypeValue(f.Type, value)
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
			values[i] = reflect_helpers.StringToTypeValue(f.Type, parts[i])
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

func isVec3(typeName string) bool {
	return typeName == "Vec3"
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
