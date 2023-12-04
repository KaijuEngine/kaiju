//go:build !F64

package matrix

import "math"

type Float = float32

const FloatSmallestNonzero = math.SmallestNonzeroFloat32

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
