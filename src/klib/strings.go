/******************************************************************************/
/* strings.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	_         = iota
	kb uint64 = 1 << (10 * iota)
	mb
	gb
	tb
)

var (
	snakeCaseMatchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	snakeCaseMatchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func ReplaceStringRecursive(s string, old string, new string) string {
	for strings.Contains(s, old) {
		s = strings.Replace(s, old, new, -1)
	}
	return s
}

func FormatFloatToNDecimals[T float32 | float64](f T, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	s := fmt.Sprintf(format, f)
	return StripFloatStringZeros(s)
}

func StripFloatStringZeros(fString string) string {
	fString = strings.TrimRight(fString, "0")
	return strings.TrimSuffix(fString, ".")
}

func ToSnakeCase(str string) string {
	snake := snakeCaseMatchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = snakeCaseMatchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func ByteCountToString(bytes uint64) string {
	switch {
	case bytes >= tb:
		return fmt.Sprintf("%.2fTB", float64(bytes)/float64(tb))
	case bytes >= gb:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// StringValueCompare compares two strings as integers or floats if possible,
// otherwise compares as strings
func StringValueCompare(a, b string) int {
	if ai, err := strconv.Atoi(a); err == nil {
		if bi, err := strconv.Atoi(b); err == nil {
			ri := ai - bi
			if ri < 0 {
				return -1
			} else if ri > 0 {
				return 1
			}
			return 0
		}
	}
	if af, err := strconv.ParseFloat(a, 64); err == nil {
		if bf, err := strconv.ParseFloat(b, 64); err == nil {
			rf := af - bf
			if rf < 0 {
				return -1
			} else if rf > 0 {
				return 1
			}
			return 0
		}
	}
	return strings.Compare(a, b)
}

func CleanNumString(v string) string {
	v = strings.TrimSpace(v)
	switch v {
	case "", "-", "-.", ".":
		return "0"
	}
	return v
}
