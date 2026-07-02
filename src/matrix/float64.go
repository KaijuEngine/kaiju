//go:build F64

/******************************************************************************/
/* float64.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package matrix

import "math"

type Float = float64

const FloatSmallestNonzero = Float(math.SmallestNonzeroFloat64)
const FloatMax = Float(math.MaxFloat64)

func Abs[T tNumber](x T) T {
	return T(math.Abs(float64(x)))
}
