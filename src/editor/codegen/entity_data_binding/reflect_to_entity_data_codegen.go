/******************************************************************************/
/* reflect_to_entity_data_codegen.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package entity_data_binding

import (
	"reflect"

	"kaijuengine.com/editor/codegen"
	"kaijuengine.com/editor/codegen/reflect_helpers"
)

func ToDataBinding(name string, target any) EntityDataEntry {
	var g EntityDataEntry
	g.Name = name
	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	t := v.Type()
	g.Fields = make([]EntityDataField, t.NumField())
	g.BoundData = target
	for i := range t.NumField() {
		f := t.Field(i)
		fieldValue := v.Field(i)
		g.Fields[i] = EntityDataField{
			Idx:  i,
			Name: f.Name,
			Type: reflect_helpers.CanonicalTypeName(f.Type.Name()),
			Pkg:  f.Type.PkgPath(),
		}
		if fieldValue.IsValid() && fieldValue.CanInterface() {
			g.Fields[i].Value = fieldValue.Interface()
		}
		g.Gen.FieldGens = append(g.Gen.FieldGens, codegen.GeneratedType{})
	}
	return g
}
