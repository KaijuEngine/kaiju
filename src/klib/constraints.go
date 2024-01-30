package klib

import "cmp"

type AnyFloat interface {
	~float32 | ~float64
}

func Clamp[T cmp.Ordered](current, minimum, maximum T) T {
	return max(minimum, min(maximum, current))
}
