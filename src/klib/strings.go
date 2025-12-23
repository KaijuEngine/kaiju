/******************************************************************************/
/* strings.go                                                                 */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package klib

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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
