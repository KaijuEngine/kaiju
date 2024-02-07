package klib

import "strings"

func ReplaceStringRecursive(s string, old string, new string) string {
	for strings.Contains(s, old) {
		s = strings.Replace(s, old, new, -1)
	}
	return s
}
