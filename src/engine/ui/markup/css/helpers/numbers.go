/******************************************************************************/
/* numbers.go                                                                 */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package helpers

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

type WindowDimensions interface {
	DotsPerMillimeter() float64
	Width() int
	Height() int
}

var arithmeticMap = map[string]func(int, int) int{
	"+": func(a, b int) int { return a + b },
	"-": func(a, b int) int { return a - b },
	"*": func(a, b int) int { return a * b },
	"/": func(a, b int) int { return a / b },
}

func ChangeNToChildCount(args []string, count int) {
	for i := range args {
		if args[i] == "n" {
			args[i] = strconv.Itoa(count)
		}
	}
}

func ArithmeticString(args []string) (int, error) {
	if len(args) == 1 {
		return strconv.Atoi(args[0])
	} else if len(args) == 2 {
		// Expected to be something like -5
		return strconv.Atoi(args[0] + args[1])
	} else {
		do := arithmeticMap["+"]
		value := 0
		negate := false
		for i := range args {
			if args[i] == "-" {
				negate = true
				continue
			} else if v, err := strconv.Atoi(args[i]); err == nil {
				if negate {
					v = -v
				}
				value = do(value, v)
			} else if f, ok := arithmeticMap[args[i]]; ok {
				do = f
			} else {
				return 0, fmt.Errorf("invalid arithmetic operator: %s", args[i])
			}
		}
		return value, nil
	}
}

func NumFromLengthWithFont(str string, window WindowDimensions, fontSize matrix.Float) matrix.Float {
	dpmm := window.DotsPerMillimeter()
	parse := func(raw string, cut int) matrix.Float {
		// strconv.ParseFloat is orders of magnitude faster than fmt.Sscanf("%f"),
		// which dominated CPU: this runs for every CSS length on every layout
		// pass. Preserve fmt's leniency (take the longest leading numeric prefix,
		// tolerating trailing garbage) by trimming trailing chars until it parses;
		// valid input parses on the first try so the common path stays fast.
		s := strings.TrimSpace(raw[:len(raw)-cut])
		for len(s) > 0 {
			if f, err := strconv.ParseFloat(s, 32); err == nil {
				if math.IsNaN(f) || math.IsInf(f, 0) {
					return 0
				}
				return matrix.Float(f)
			}
			s = s[:len(s)-1]
		}
		return 0
	}
	switch {
	case strings.HasSuffix(str, "vmin"):
		size := parse(str, 4)
		w := matrix.Float(window.Width())
		h := matrix.Float(window.Height())
		if h < w {
			w = h
		}
		return w * (size / 100)
	case strings.HasSuffix(str, "vmax"):
		size := parse(str, 4)
		w := matrix.Float(window.Width())
		h := matrix.Float(window.Height())
		if h > w {
			w = h
		}
		return w * (size / 100)
	case strings.HasSuffix(str, "rem"):
		size := parse(str, 3)
		// Root font size support is not yet wired through style inheritance.
		// For now rem is based on the engine default root em size.
		return size * rendering.DefaultFontEMSize
	case strings.HasSuffix(str, "vw"):
		size := parse(str, 2)
		return matrix.Float(window.Width()) * (size / 100)
	case strings.HasSuffix(str, "vh"):
		size := parse(str, 2)
		return matrix.Float(window.Height()) * (size / 100)
	case strings.HasSuffix(str, "ch"):
		size := parse(str, 2)
		// Approximation until font metric support is available:
		// 1ch ~= 0.5em
		return size * fontSize * 0.5
	case strings.HasSuffix(str, "px"),
		strings.HasSuffix(str, "em"),
		strings.HasSuffix(str, "ex"),
		strings.HasSuffix(str, "cm"),
		strings.HasSuffix(str, "mm"),
		strings.HasSuffix(str, "in"),
		strings.HasSuffix(str, "pt"),
		strings.HasSuffix(str, "pc"):
		size := parse(str, 2)
		switch str[len(str)-2:] {
		case "px":
			return size
		case "em", "ex":
			return size * fontSize
		case "cm":
			return matrix.Float(dpmm) * matrix.Float(size*10)
		case "mm":
			return matrix.Float(dpmm) * size
		case "in":
			return matrix.Float(dpmm) * matrix.Float(size*25.4)
		case "pt":
			return matrix.Float(dpmm) * matrix.Float(size*25.4/72)
		case "pc":
			return matrix.Float(dpmm) * matrix.Float(size*25.4/6)
		}
	case strings.HasSuffix(str, "%"):
		size := parse(str, 1)
		return size / 100
	}
	return 0
}

// NumFromLength resolves CSS lengths with the default font size context.
// For properties that depend on the current element font, use NumFromLengthWithFont.
func NumFromLength(str string, window WindowDimensions) matrix.Float {
	return NumFromLengthWithFont(str, window, rendering.DefaultFontEMSize)
}
