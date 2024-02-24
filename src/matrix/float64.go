//go:build F64

/******************************************************************************/
/* float64.go                                                                 */
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

package matrix

import "math"

type Float = float64

const FloatSmallestNonzero = math.SmallestNonzeroFloat64
const FloatMax = math.MaxFloat64

func Abs(x Float) Float {
	return math.Abs(x)
}

func Min(a Float, b Float) Float {
	return math.Min(a, b)
}

func Max(a Float, b Float) Float {
	return math.Max(a, b)
}

func Acos(x Float) Float {
	return math.Acos(x)
}

func Sqrt(x Float) Float {
	return math.Sqrt(x)
}

func Log2(x Float) Float {
	return math.Log2(x)
}

func Floor(x Float) Float {
	return math.Floor(x)
}

func Ceil(x Float) Float {
	return math.Ceil(x)
}

func Sin(x Float) Float {
	return math.Sin(x)
}

func Cos(x Float) Float {
	return math.Cos(x)
}

func Tan(x Float) Float {
	return math.Tan(x)
}

func Asin(x Float) Float {
	return math.Asin(x)
}

func Atan(x Float) Float {
	return math.Atan(x)
}

func Atan2(y Float, x Float) Float {
	return math.Atan2(y, x)
}

func Pow(x Float, y Float) Float {
	return math.Pow(x, y)
}

func IsNan(x Float) bool {
	return math.IsNaN(x)
}

func IsInf(x Float, sign int) bool {
	return math.IsInf(x, sign)
}

func Inf(sign int) Float {
	return math.Inf(sign)
}
