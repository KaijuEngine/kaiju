package codegen

import (
	"encoding/gob"
	"reflect"
)

type RuntimeType struct {
	Generator *GeneratedType
	Value     reflect.Value
}

func (g *GeneratedType) New() RuntimeType {
	rt := RuntimeType{
		Generator: g,
		Value:     reflect.New(g.Type),
	}
	if !g.registered {
		name := "*" + g.PkgPath + "." + g.Name
		gob.UnRegisterName(name)
		gob.RegisterNamedType(name, rt.Value.Type())
		g.registered = true
	}
	return rt
}
