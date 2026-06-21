/******************************************************************************/
/* float.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import "math"

const Tiny = 0.0001
const Roughly = 0.001

type floatInput interface {
	~float32 | ~float64 | ~int
}

func Min[T1, T2 floatInput](a T1, b T2) Float {
	return Float(math.Min(float64(a), float64(b)))
}

func Max[T1, T2 floatInput](a T1, b T2) Float {
	return Float(math.Max(float64(a), float64(b)))
}

func Acos[T floatInput](x T) Float {
	return Float(math.Acos(float64(x)))
}

func Sqrt[T floatInput](x T) Float {
	return Float(math.Sqrt(float64(x)))
}

func Log2[T floatInput](x T) Float {
	return Float(math.Log2(float64(x)))
}

func Floor[T floatInput](x T) Float {
	return Float(math.Floor(float64(x)))
}

func Ceil[T floatInput](x T) Float {
	return Float(math.Ceil(float64(x)))
}

func Sin[T floatInput](x T) Float {
	return Float(math.Sin(float64(x)))
}

func Cos[T floatInput](x T) Float {
	return Float(math.Cos(float64(x)))
}

func Tan[T floatInput](x T) Float {
	return Float(math.Tan(float64(x)))
}

func Asin[T floatInput](x T) Float {
	return Float(math.Asin(float64(x)))
}

func Atan[T floatInput](x T) Float {
	return Float(math.Atan(float64(x)))
}

func Atan2[T1, T2 floatInput](y T1, x T2) Float {
	return Float(math.Atan2(float64(y), float64(x)))
}

func Pow[T1, T2 floatInput](x T1, y T2) Float {
	return Float(math.Pow(float64(x), float64(y)))
}

func IsNaN[T floatInput](x T) bool {
	return math.IsNaN(float64(x))
}

func IsInf[T floatInput](x T, sign int) bool {
	return math.IsInf(float64(x), sign)
}

func Inf(sign int) Float {
	return Float(math.Inf(sign))
}

func NaN() Float {
	return Float(math.NaN())
}

func Mod[T1, T2 floatInput](x T1, y T2) Float {
	return Float(math.Mod(float64(x), float64(y)))
}

func Round[T floatInput](x T) Float {
	return Float(math.Round(float64(x)))
}

func Lerp[T1, T2, T3 floatInput](v0 T1, v1 T2, t T3) Float {
	return (1-Float(t))*Float(v0) + Float(t)*Float(v1)
}
