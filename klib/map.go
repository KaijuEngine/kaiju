package klib

func ClearMap[K comparable, V any](src map[K]V) {
	for k := range src {
		delete(src, k)
	}
}
