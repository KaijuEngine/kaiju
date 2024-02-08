package document

import (
	"kaiju/matrix"
	"regexp"
	"strconv"
	"strings"
)

func asFloat(valStr string) float32 {
	if len(valStr) == 0 {
		valStr = "0"
	}
	v, _ := strconv.ParseFloat(valStr, 32)
	return float32(v)
}

func convertHex(hex string) [4]int {
	out := [4]int{0, 0, 0, 255}
	if strings.HasPrefix(hex, "#") {
		hex = hex[1:]
	}
	if len(hex) == 3 {
		hex = hex[0:1] + hex[0:1] + hex[1:2] + hex[1:2] + hex[2:3] + hex[2:3]
	}
	var re *regexp.Regexp = nil
	if len(hex) == 8 {
		re = regexp.MustCompile(`([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})`)
	} else if len(hex) == 6 {
		re = regexp.MustCompile(`([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})`)
	}
	if re != nil {
		matches := re.FindStringSubmatch(hex)
		for i := 0; i < len(matches)-1; i++ {
			v, _ := strconv.ParseInt(matches[i+1], 16, 32)
			out[i] = int(v)
		}
	}
	return out
}

func convertColor(color string) matrix.Color {
	outColor := matrix.ColorWhite()
	if strings.HasPrefix(color, "rgba(") {
		re := regexp.MustCompile(`rgba\((\d+),(\d+),(\d+),(\d+)\)`)
		matches := re.FindStringSubmatch(color)
		if len(matches) == 5 {
			r, _ := strconv.Atoi(matches[1])
			g, _ := strconv.Atoi(matches[2])
			b, _ := strconv.Atoi(matches[3])
			a, _ := strconv.Atoi(matches[4])
			outColor = matrix.ColorRGBAInt(r, g, b, a)
		}
	} else if strings.HasPrefix(color, "rgb(") {
		re := regexp.MustCompile(`rgb\((\d+),(\d+),(\d+)\)`)
		matches := re.FindStringSubmatch(color)
		if len(matches) == 4 {
			r, _ := strconv.Atoi(matches[1])
			g, _ := strconv.Atoi(matches[2])
			b, _ := strconv.Atoi(matches[3])
			outColor = matrix.ColorRGBInt(r, g, b)
		}
	} else if strings.HasPrefix(color, "#") {
		rgba := convertHex(color)
		outColor = matrix.ColorRGBAInt(rgba[0], rgba[1], rgba[2], rgba[3])
	}
	return outColor
}
