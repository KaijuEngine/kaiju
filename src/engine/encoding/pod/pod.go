/******************************************************************************/
/* pod.go                                                                     */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package pod

import (
	"fmt"
	"kaijuengine.com/engine/collision"
	"kaijuengine.com/matrix"
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
