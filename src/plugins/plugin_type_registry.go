/******************************************************************************/
/* plugin_type_registry.go                                                    */
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

package plugins

import (
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"reflect"
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
