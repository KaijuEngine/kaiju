package klib

import (
	"log"
	"math"
)

func Must(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
		panic(err)
	}
}

func MustReturn[T any](ret T, err error) T {
	if err != nil {
		log.Printf("Error: %v", err)
		panic(err)
	}
	return ret
}

func Should(err error) bool {
	log.Printf("Error: %v", err)
	return err != nil
}

func ShouldReturn[T any](ret T, err error) T {
	if err != nil {
		log.Printf("Error: %v", err)
	}
	return ret
}

func FloatEquals[T AnyFloat](a, b T) bool {
	return math.Abs(float64(a-b)) < math.SmallestNonzeroFloat64
}
