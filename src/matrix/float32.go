//go:build !F64

/******************************************************************************/
/* float32.go                                                                 */
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

import "math"

type Float = float32

const FloatSmallestNonzero = Float(math.SmallestNonzeroFloat32)
const FloatMax = Float(math.MaxFloat32)

func Abs(x Float) Float {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}

func Min(a Float, b Float) Float {
	return Float(math.Min(float64(a), float64(b)))
}

func Max(a Float, b Float) Float {
	return Float(math.Max(float64(a), float64(b)))
}

func Acos(x Float) Float {
	return Float(math.Acos(float64(x)))
}

func Sqrt(x Float) Float {
	return Float(math.Sqrt(float64(x)))
}

func Log2(x Float) Float {
	return Float(math.Log2(float64(x)))
}

func Floor(x Float) Float {
	return Float(math.Floor(float64(x)))
}

func Ceil(x Float) Float {
	return Float(math.Ceil(float64(x)))
}

func Sin(x Float) Float {
	return Float(math.Sin(float64(x)))
}

func Cos(x Float) Float {
	return Float(math.Cos(float64(x)))
}

func Tan(x Float) Float {
	return Float(math.Tan(float64(x)))
}

func Asin(x Float) Float {
	return Float(math.Asin(float64(x)))
}

func Atan(x Float) Float {
	return Float(math.Atan(float64(x)))
}

func Atan2(y Float, x Float) Float {
	return Float(math.Atan2(float64(y), float64(x)))
}

func Pow(x Float, y Float) Float {
	return Float(math.Pow(float64(x), float64(y)))
}

func IsNaN(x Float) bool {
	return math.IsNaN(float64(x))
}

func IsInf(x Float, sign int) bool {
	return math.IsInf(float64(x), sign)
}

func Inf(sign int) Float {
	return Float(math.Inf(sign))
}

func NaN() Float {
	return Float(math.NaN())
}

func Mod(x Float, y Float) Float {
	return Float(math.Mod(float64(x), float64(y)))
}

func Round(x Float) Float {
	return Float(math.Round(float64(x)))
}

func Lerp(v0, v1, t Float) Float {
	return (1-t)*v0 + t*v1
}
