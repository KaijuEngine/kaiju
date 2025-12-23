/******************************************************************************/
/* matrix_config.go                                                           */
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

package matrix

import (
	"math"
)

type ColorComponent = int
type VectorComponent = int
type QuaternionComponent = int

const (
	R ColorComponent = iota
	G
	B
	A
)

const (
	Vx VectorComponent = iota
	Vy
	Vz
	Vw
)

const (
	Qw QuaternionComponent = iota
	Qx
	Qy
	Qz
)

type tFloatingPoint interface {
	~float32 | ~float64
}

type tSigned interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type tUnsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type tInteger interface {
	tSigned | tUnsigned
}

type tNumber interface {
	tInteger | tFloatingPoint
}

type tVector interface {
	Vec2 | Vec3 | Vec4 | Quaternion
}

type tMatrix interface {
	Mat3 | Mat4
}

const RadToDegVal = (180.0 / math.Pi)
const DegToRadVal = (math.Pi / 180.0)

func Rad2Deg(radian Float) Float {
	return radian * (180.0 / math.Pi)
}

func Deg2Rad(degree Float) Float {
	return degree * (math.Pi / 180.0)
}

func Approx(a, b Float) bool {
	return math.Abs(float64(a-b)) < FloatSmallestNonzero
}

func ApproxTo(a, b, tolerance Float) bool {
	return math.Abs(float64(a-b)) < float64(tolerance)
}

func Clamp(current, minimum, maximum Float) Float {
	return max(minimum, min(maximum, current))
}

func AbsInt(a int) int { return a & int(^uint(0)>>1) }
