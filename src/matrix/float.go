/******************************************************************************/
/* float.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import "math"

const Tiny = 0.0001
const Roughly = 0.001

func Acos[T tNumber](x T) T {
	return T(math.Acos(float64(x)))
}

func Sqrt[T tNumber](x T) T {
	return T(math.Sqrt(float64(x)))
}

func Log2[T tNumber](x T) T {
	return T(math.Log2(float64(x)))
}

func Floor[T tNumber](x T) T {
	return T(math.Floor(float64(x)))
}

func Ceil[T tNumber](x T) T {
	return T(math.Ceil(float64(x)))
}

func Sin[T tNumber](x T) T {
	return T(math.Sin(float64(x)))
}

func Cos[T tNumber](x T) T {
	return T(math.Cos(float64(x)))
}

func Tan[T tNumber](x T) T {
	return T(math.Tan(float64(x)))
}

func Asin[T tNumber](x T) T {
	return T(math.Asin(float64(x)))
}

func Atan[T tNumber](x T) T {
	return T(math.Atan(float64(x)))
}

func Atan2[T tNumber](y T, x T) T {
	return T(math.Atan2(float64(y), float64(x)))
}

func Pow[T tNumber](x T, y T) T {
	return T(math.Pow(float64(x), float64(y)))
}

func IsNaN[T tNumber](x T) bool {
	return math.IsNaN(float64(x))
}

func IsInf[T tNumber](x T, sign int) bool {
	return math.IsInf(float64(x), sign)
}

func Inf(sign int) Float {
	return Float(math.Inf(sign))
}

func NaN() Float {
	return Float(math.NaN())
}

func Mod[T tNumber](x T, y T) T {
	return T(math.Mod(float64(x), float64(y)))
}

func Round[T tNumber](x T) T {
	return T(math.Round(float64(x)))
}

func Lerp[T tNumber](v0 T, v1 T, t T) T {
	return T((1-float64(t))*float64(v0) + float64(t)*float64(v1))
}
