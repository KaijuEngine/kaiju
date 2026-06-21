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

func Abs[T floatInput](x T) Float {
	return math.Abs(float64(x))
}
