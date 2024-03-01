package codegen

import (
	"kaiju/engine"
	"kaiju/matrix"
	"reflect"
)

var (
	registry = make(map[string]reflect.Type)
)

func init() {
	RegisterTypeName("matrix.Float", matrix.Float(0))
	RegisterType(matrix.Color{})
	RegisterType(matrix.Color{})
	RegisterType(matrix.Mat3{})
	RegisterType(matrix.Mat3{})
	RegisterType(matrix.Mat4{})
	RegisterType(matrix.Mat4{})
	RegisterType(matrix.Quaternion{})
	RegisterType(matrix.Quaternion{})
	RegisterType(matrix.Transform{})
	RegisterType(matrix.Transform{})
	RegisterType(matrix.Vec2{})
	RegisterType(matrix.Vec2{})
	RegisterType(matrix.Vec2i{})
	RegisterType(matrix.Vec2i{})
	RegisterType(matrix.Vec3{})
	RegisterType(matrix.Vec3{})
	RegisterType(matrix.Vec3i{})
	RegisterType(matrix.Vec3i{})
	RegisterType(matrix.Vec4{})
	RegisterType(matrix.Vec4{})
	RegisterType(matrix.Vec4i{})
	RegisterType(engine.Entity{})
	RegisterType(engine.Host{})
}

func RegisterType(t any) {
	registry[reflect.TypeOf(t).String()] = reflect.TypeOf(t)
}

func RegisterTypeName(name string, t any) {
	registry[name] = reflect.TypeOf(t)
}
