package codegen

import "reflect"

type GeneratedType struct {
	Pkg        string
	PkgPath    string
	Name       string
	Fields     []reflect.StructField
	Type       reflect.Type
	registered bool
}
