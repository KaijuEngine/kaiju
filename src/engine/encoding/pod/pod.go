package pod

import (
	"fmt"
	"kaiju/engine/collision"
	"kaiju/matrix"
	"reflect"
	"sync"
)

const (
	kindTypeSliceArray = uint8(0xFF)
	kindTypeMap        = uint8(0xFE)
	kindTypeLimit      = kindTypeMap
	// 0x00 - 0xFD are reserved for the registration keys
)

var (
	registry = sync.Map{}
)

func init() {
	Register(bool(false))
	Register(int(0))
	Register(int8(0))
	Register(int16(0))
	Register(int32(0))
	Register(int64(0))
	Register(uint8(0))
	Register(uint16(0))
	Register(uint32(0))
	Register(uint64(0))
	Register(float32(0))
	Register(float64(0))
	Register(complex64(0))
	Register(complex128(0))
	Register(rune(0))
	Register(string(""))
	Register(matrix.Vec2{})
	Register(matrix.Vec3{})
	Register(matrix.Vec4{})
	Register(matrix.Color{})
	Register(matrix.Color8{})
	Register(matrix.Quaternion{})
	Register(matrix.Mat3{})
	Register(matrix.Mat4{})
	Register(collision.AABB{})
	Register(collision.Ray{})
	Register(collision.Frustum{})
	Register(collision.Plane{})
	Register(collision.Triangle{})
}

func Unregister(layout any) {
	registry.Delete(qualifiedName(reflect.TypeOf(layout)))
}

func UnregisterGenerated(pkg, name string) {
	registry.Delete(pkg + "." + name)
}

func Register(layout any) error {
	t := reflect.TypeOf(layout)
	q := qualifiedName(t)
	if _, ok := registry.LoadOrStore(q, t); ok {
		return fmt.Errorf("the name '%s' has already been registered in pod", q)
	}
	return nil
}

func RegisterGenerated(pkg, name string, genType reflect.Type) error {
	q := pkg + "." + name
	if _, ok := registry.LoadOrStore(q, genType); ok {
		return fmt.Errorf("the name '%s' has already been registered in pod", q)
	}
	return nil
}

func QualifiedNameForLayout(layout any) string {
	return qualifiedName(reflect.TypeOf(layout))
}
