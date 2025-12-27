/******************************************************************************/
/* emitter_path_funcs.go                                                      */
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

package vfx

import (
	"kaiju/klib"
	"kaiju/matrix"
	"log/slog"
	"math"
)

var pathFunctions = map[string]func(t float64) matrix.Vec3{}

func init() {
	RegisterPathFunc("None", nil)
	RegisterPathFunc("Circle", pathFuncCircle)
}

func RegisterPathFunc(name string, fn func(t float64) matrix.Vec3) {
	if _, ok := pathFunctions[name]; ok {
		slog.Error("there is already a path function registered with that name", "name", name)
		return
	}
	pathFunctions[name] = fn
}

func EditorReflectionOptions(name string) []string {
	switch name {
	case "PathFuncName":
		return klib.MapKeys(pathFunctions)
	default:
		return []string{}
	}
}

func pathFuncCircle(t float64) matrix.Vec3 {
	pos := matrix.Vec3{}
	for t < 0 {
		t += 1
	}
	for t > 1 {
		t -= 1
	}
	angle := matrix.Float(2 * math.Pi * t)
	pos.SetX(matrix.Cos(angle))
	pos.SetZ(matrix.Sin(angle))
	return pos
}
