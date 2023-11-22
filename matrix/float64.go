//go:build F64

package matrix

import "math"

type Float = float64

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
