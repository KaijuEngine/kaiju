//go:build F64

/******************************************************************************/
/* float64.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
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

func IsNaN(x Float) bool {
	return math.IsNaN(x)
}

func IsInf(x Float, sign int) bool {
	return math.IsInf(x, sign)
}

func Inf(sign int) Float {
	return math.Inf(sign)
}

func NaN() Float {
	return math.NaN()
}

func Mod(x Float, y Float) Float {
	return math.Mod(x, y)
}

func Round(x Float) Float {
	return math.Round(x)
}

func Lerp(v0, v1, t Float) Float {
	return (1-t)*v0 + t*v1
}
