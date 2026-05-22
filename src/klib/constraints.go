/******************************************************************************/
/* constraints.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import "cmp"

type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface {
	SignedInteger | Unsigned
}

type Float interface {
	~float32 | ~float64
}

type Signed interface {
	SignedInteger | Float
}

type Complex interface {
	~complex64 | ~complex128
}

type Ordered interface {
	Integer | Float | ~string
}

type Number interface {
	Integer | Float | Complex
}

func Clamp[T cmp.Ordered](current, minimum, maximum T) T {
	return max(minimum, min(maximum, current))
}

func ClampAbs[T Signed](value, minimum T) T {
	if value > minimum {
		return minimum
	}
	if value < -minimum {
		return -minimum
	}
	return value
}
