/******************************************************************************/
/* plugin_type_registry.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package plugins

import (
	"reflect"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

var (
	GamePluginRegistry = []reflect.Type{}
)

func reflectedTypeName(t reflect.Type) string {
	switch t {
	case reflect.TypeFor[matrix.Vec2]():
		return "Vec2"
	case reflect.TypeFor[matrix.Vec3]():
		return "Vec3"
	case reflect.TypeFor[matrix.Vec4]():
		return "Vec4"
	default:
		return t.Name()
	}
}

func reflectedTypes() []reflect.Type {
	defer tracing.NewRegion("plugins.reflectedTypes").End()
	return append([]reflect.Type{
		reflect.TypeFor[matrix.Vec2](),
		reflect.TypeFor[matrix.Vec2i](),
		reflect.TypeFor[matrix.Vec3](),
		reflect.TypeFor[matrix.Vec3i](),
		reflect.TypeFor[matrix.Vec4](),
		reflect.TypeFor[matrix.Vec4i](),
		reflect.TypeFor[matrix.Quaternion](),
		reflect.TypeFor[matrix.Mat3](),
		reflect.TypeFor[matrix.Mat4](),
	}, GamePluginRegistry...)
}
