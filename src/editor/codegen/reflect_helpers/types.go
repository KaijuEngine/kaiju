/******************************************************************************/
/* types.go                                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package reflect_helpers

import (
	"log/slog"
	"strconv"
	"strings"
	"unsafe"

	"kaijuengine.com/klib"
	"kaijuengine.com/matrix"
)

func bitSize[T klib.Number]() int { return int(unsafe.Sizeof(T(0))) * 8 }

// CanonicalTypeName maps the runtime names of the generic matrix vector
// implementations back to the public aliases used in entity-data source.
// reflect reports matrix.Vec3 as "Vec3T[float32]", for example, because Vec3
// is an alias. The editor's field classification and default parser should
// continue to treat that type as "Vec3".
func CanonicalTypeName(typeName string) string {
	for _, vectorName := range []string{"Vec2", "Vec3", "Vec4"} {
		if strings.HasPrefix(typeName, vectorName+"T[") && strings.HasSuffix(typeName, "]") {
			return vectorName
		}
	}
	return typeName
}

func StringToTypeValue(typeName, v string) any {
	typeName = CanonicalTypeName(typeName)
	switch typeName {
	case "string":
		return v
	case "bool":
		switch strings.ToLower(v) {
		case "false":
			return false
		case "true":
			return true
		default:
			slog.Warn("unexpected tag string value for bool, expected 'true' or 'false'", "value", v)
			return true
		}
	case "int":
		v = klib.CleanNumString(v)
		return int(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int]())))
	case "int8":
		v = klib.CleanNumString(v)
		return int8(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int8]())))
	case "int16":
		v = klib.CleanNumString(v)
		return int16(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int16]())))
	case "int32":
		v = klib.CleanNumString(v)
		return int32(klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int32]())))
	case "int64":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseInt(v, 0, bitSize[int64]()))
	case "uint":
		v = klib.CleanNumString(v)
		return uint(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint]())))
	case "uint8":
		v = klib.CleanNumString(v)
		return uint8(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint8]())))
	case "uint16":
		v = klib.CleanNumString(v)
		return uint16(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint16]())))
	case "uint32":
		v = klib.CleanNumString(v)
		return uint32(klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint32]())))
	case "uint64":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uint64]()))
	case "float32":
		v = klib.CleanNumString(v)
		return matrix.Float(klib.ShouldReturn(strconv.ParseFloat(v, bitSize[matrix.Float]())))
	case "float64":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseFloat(v, bitSize[float64]()))
	case "uintptr":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseUint(v, 0, bitSize[uintptr]()))
	case "complex64":
		v = klib.CleanNumString(v)
		return complex64(klib.ShouldReturn(strconv.ParseComplex(v, bitSize[complex64]())))
	case "complex128":
		v = klib.CleanNumString(v)
		return klib.ShouldReturn(strconv.ParseComplex(v, bitSize[complex128]()))
	case "Vec2":
		out := matrix.Vec2{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := klib.CleanNumString(parts[i])
			out[i] = matrix.Float(klib.ShouldReturn(strconv.ParseFloat(p, bitSize[matrix.Float]())))
		}
		return out
	case "Vec3":
		out := matrix.Vec3{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := klib.CleanNumString(parts[i])
			out[i] = matrix.Float(klib.ShouldReturn(strconv.ParseFloat(p, bitSize[matrix.Float]())))
		}
		return out
	case "Vec4":
		out := matrix.Vec4{}
		parts := strings.Split(v, ",")
		for i := range min(len(out), len(parts)) {
			p := klib.CleanNumString(parts[i])
			out[i] = matrix.Float(klib.ShouldReturn(strconv.ParseFloat(p, bitSize[matrix.Float]())))
		}
		return out
	}
	return nil
}
