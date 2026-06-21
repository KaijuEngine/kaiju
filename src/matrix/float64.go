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

func Abs[T tNumber](x T) Float {
	return math.Abs(float64(x))
}

func Min[T1, T2 tNumber](a T1, b T2) Float {
	return math.Min(float64(a), float64(b))
}

func Max[T1, T2 tNumber](a T1, b T2) Float {
	return math.Max(float64(a), float64(b))
}

func Acos[T tNumber](x T) Float {
	return math.Acos(float64(x))
}

func Sqrt[T tNumber](x T) Float {
	return math.Sqrt(float64(x))
}

func Log2[T tNumber](x T) Float {
	return math.Log2(float64(x))
}

func Floor[T tNumber](x T) Float {
	return math.Floor(float64(x))
}

func Ceil[T tNumber](x T) Float {
	return math.Ceil(float64(x))
}

func Sin[T tNumber](x T) Float {
	return math.Sin(float64(x))
}

func Cos[T tNumber](x T) Float {
	return math.Cos(float64(x))
}

func Tan[T tNumber](x T) Float {
	return math.Tan(float64(x))
}

func Asin[T tNumber](x T) Float {
	return math.Asin(float64(x))
}

func Atan[T tNumber](x T) Float {
	return math.Atan(float64(x))
}

func Atan2[T1, T2 tNumber](y T1, x T2) Float {
	return math.Atan2(float64(y), float64(x))
}

func Pow[T1, T2 tNumber](x T1, y T2) Float {
	return math.Pow(float64(x), float64(y))
}

func IsNaN[T tNumber](x T) bool {
	return math.IsNaN(float64(x))
}

func IsInf[T tNumber](x T, sign int) bool {
	return math.IsInf(float64(x), sign)
}

func Inf(sign int) Float {
	return math.Inf(sign)
}

func NaN() Float {
	return math.NaN()
}

func Mod[T1, T2 tNumber](x T1, y T2) Float {
	return math.Mod(float64(x), float64(y))
}

func Round[T tNumber](x T) Float {
	return math.Round(float64(x))
}

func Lerp[T1, T2, T3 tNumber](v0 T1, v1 T2, t T3) Float {
	return (1-float64(t))*float64(v0) + float64(t)*float64(v1)
}
