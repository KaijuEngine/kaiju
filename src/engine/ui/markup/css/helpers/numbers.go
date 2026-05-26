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

func NumFromLengthWithFont(str string, window WindowDimensions, fontSize float32) float32 {
	dpmm := window.DotsPerMillimeter()
	parse := func(raw string, cut int) float32 {
		var v float32
		fmt.Sscanf(raw[:len(raw)-cut], "%f", &v)
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return 0
		}
		return v
	}
	switch {
	case strings.HasSuffix(str, "vmin"):
		size := parse(str, 4)
		w := float32(window.Width())
		h := float32(window.Height())
		if h < w {
			w = h
		}
		return w * (size / 100)
	case strings.HasSuffix(str, "vmax"):
		size := parse(str, 4)
		w := float32(window.Width())
		h := float32(window.Height())
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
		return float32(window.Width()) * (size / 100)
	case strings.HasSuffix(str, "vh"):
		size := parse(str, 2)
		return float32(window.Height()) * (size / 100)
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
			return float32(dpmm) * float32(size*10)
		case "mm":
			return float32(dpmm) * size
		case "in":
			return float32(dpmm) * float32(size*25.4)
		case "pt":
			return float32(dpmm) * float32(size*25.4/72)
		case "pc":
			return float32(dpmm) * float32(size*25.4/6)
		}
	case strings.HasSuffix(str, "%"):
		size := parse(str, 1)
		return size / 100
	}
	return 0
}

// NumFromLength resolves CSS lengths with the default font size context.
// For properties that depend on the current element font, use NumFromLengthWithFont.
func NumFromLength(str string, window WindowDimensions) float32 {
	return NumFromLengthWithFont(str, window, rendering.DefaultFontEMSize)
}
