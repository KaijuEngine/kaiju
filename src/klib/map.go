/******************************************************************************/
/* map.go                                                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"cmp"
	"slices"
)

func MapKeys[T comparable, U any](m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func MapKeysSorted[T cmp.Ordered, U any](m map[T]U) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func MapValues[T comparable, U any](m map[T]U) []U {
	values := make([]U, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func MapJoin[T comparable, U any](a, b map[T]U) map[T]U {
	if len(b) < len(a) {
		a, b = b, a
	}
	for k, v := range b {
		a[k] = v
	}
	return a
}
