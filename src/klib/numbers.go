package klib

import "strings"

func CleanNumString(v string) string {
	v = strings.TrimSpace(v)
	switch v {
	case "", "-", "-.", ".":
		return "0"
	}
	return v
}
