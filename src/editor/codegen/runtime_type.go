/******************************************************************************/
/* runtime_type.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package codegen

import (
	"reflect"

	"kaijuengine.com/engine/encoding/pod"
	"kaijuengine.com/platform/profiler/tracing"
)

type RuntimeType struct {
	Generator *GeneratedType
	Value     reflect.Value
}

func (g *GeneratedType) New() RuntimeType {
	defer tracing.NewRegion("GeneratedType.New").End()
	rt := RuntimeType{
		Generator: g,
		Value:     reflect.New(g.Type),
	}
	if !g.registered {
		pod.UnregisterGenerated(g.Pkg, g.Name)
		pod.RegisterGenerated(g.Pkg, g.Name, rt.Value.Elem().Type())
		g.registered = true
	}
	return rt
}
