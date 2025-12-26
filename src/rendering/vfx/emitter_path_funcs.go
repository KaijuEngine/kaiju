package vfx

import (
	"kaiju/klib"
	"kaiju/matrix"
	"log/slog"
	"math"
)

var pathFunctions = map[string]func(t float64) matrix.Vec3{}

func init() {
	RegisterPathFunc("None", nil)
	RegisterPathFunc("Circle", pathFuncCircle)
}

func RegisterPathFunc(name string, fn func(t float64) matrix.Vec3) {
	if _, ok := pathFunctions[name]; ok {
		slog.Error("there is already a path function registered with that name", "name", name)
		return
	}
	pathFunctions[name] = fn
}

func EditorReflectionOptions(name string) []string {
	switch name {
	case "PathFuncName":
		return klib.MapKeys(pathFunctions)
	default:
		return []string{}
	}
}

func pathFuncCircle(t float64) matrix.Vec3 {
	pos := matrix.Vec3{}
	for t < 0 {
		t += 1
	}
	for t > 1 {
		t -= 1
	}
	angle := matrix.Float(2 * math.Pi * t)
	pos.SetX(matrix.Cos(angle))
	pos.SetZ(matrix.Sin(angle))
	return pos
}
