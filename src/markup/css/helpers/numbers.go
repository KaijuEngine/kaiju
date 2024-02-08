package helpers

import (
	"fmt"
	"kaiju/klib"
	"kaiju/windowing"
	"strconv"
	"strings"
)

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

func NumFromLengthWithFont(str string, window *windowing.Window, fontSize float32) float32 {
	w := window.Width()
	//h := window.Height
	//wmm, hmm, _ := window.GetDPI()
	wmm, _, _ := window.GetDPI()
	var suffix string
	if str[len(str)-1] == '%' {
		suffix = "%"
		str = str[:len(str)-1]
	} else if len(str) > 2 {
		validSuffixes := []string{"px", "em", "ex", "cm", "mm", "in", "pt", "pc"}
		valid := false
		for i := range validSuffixes {
			valid = valid || strings.HasSuffix(str, validSuffixes[i])
		}
		if valid {
			suffix = str[len(str)-2:]
			str = str[:len(str)-2]
		}
	}
	var size float32
	fmt.Sscanf(str, "%f", &size)
	switch suffix {
	case "%":
		size = size / 100
	case "px":
		// Read value is the size
	case "ex":
		// Relative to the font size of a lowercase letter like a, c, m, or o
		fallthrough
	case "em":
		size = size * fontSize
	case "cm":
		size = float32(klib.MM2PX(w, wmm, int(size*10)))
	case "mm":
		size = float32(klib.MM2PX(w, wmm, int(size)))
	case "in":
		size = float32(klib.MM2PX(w, wmm, int(size*25.4)))
	case "pt":
		size = float32(klib.MM2PX(w, wmm, int(size*25.4/72)))
	case "pc":
		size = float32(klib.MM2PX(w, wmm, int(size*25.4/6)))
	default:
		size = 0
	}
	return size
}

func NumFromLength(str string, window *windowing.Window) float32 {
	return NumFromLengthWithFont(str, window, 0)
}