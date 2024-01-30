package klib

import "math/rand"

type Anything interface {
}

func RemoveOrdered[T Anything](slice []T, idx int) []T {
	return append(slice[:idx], slice[idx+1:]...)
}

func RemoveUnordered[T Anything](slice []T, idx int) []T {
	slice[idx] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func Shuffle[T Anything](slice []T, rng *rand.Rand) {
	for i := len(slice) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func Contains[T comparable](slice []T, item T) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}
