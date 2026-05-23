/******************************************************************************/
/* generated_type.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package codegen

import (
	"path"
	"reflect"
)

type GeneratedType struct {
	Pkg                string
	PkgPath            string
	Name               string
	Fields             []reflect.StructField
	FieldGens          []GeneratedType
	Type               reflect.Type
	EnumValues         map[string]any
	RegisterKey        string
	registered         bool
	satisfiesInterface bool
}

func (g *GeneratedType) IsValid() bool { return g.Name != "" }

func GeneratedTypeFromValue(registerKey string, value any) GeneratedType {
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	fields := reflect.VisibleFields(t)
	return GeneratedType{
		Pkg:         path.Base(t.PkgPath()),
		PkgPath:     t.PkgPath(),
		Name:        t.Name(),
		Fields:      fields,
		FieldGens:   make([]GeneratedType, len(fields)),
		Type:        t,
		RegisterKey: registerKey,
	}
}
