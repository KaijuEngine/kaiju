/******************************************************************************/
/* codgen_registry.go                                                         */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE /* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY    */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
