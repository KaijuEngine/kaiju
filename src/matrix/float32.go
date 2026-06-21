//go:build !F64

/******************************************************************************/
/* float32.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import "math"

type Float = float32

const FloatSmallestNonzero = Float(math.SmallestNonzeroFloat32)
const FloatMax = Float(math.MaxFloat32)

func Abs(x Float) Float {
	return math.Float32frombits(math.Float32bits(x) &^ (1 << 31))
}
