package plugins

import (
	"kaiju/matrix"
	"reflect"
)

func reflectedTypes() []reflect.Type {
	return []reflect.Type{
		reflect.TypeFor[matrix.Vec2](),
		reflect.TypeFor[matrix.Vec2i](),
		reflect.TypeFor[matrix.Vec3](),
		reflect.TypeFor[matrix.Vec3i](),
		reflect.TypeFor[matrix.Vec4](),
		reflect.TypeFor[matrix.Vec4i](),
		reflect.TypeFor[matrix.Quaternion](),
		reflect.TypeFor[matrix.Mat3](),
		reflect.TypeFor[matrix.Mat4](),
	}
}
