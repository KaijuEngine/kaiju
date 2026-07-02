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

func Abs[T tNumber](x T) T {
	return T(math.Float32frombits(math.Float32bits(float32(x)) &^ (1 << 31)))
}
