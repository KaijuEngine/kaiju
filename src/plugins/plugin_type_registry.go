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
