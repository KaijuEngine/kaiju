package klib

import (
	"log/slog"
	"strconv"
	"strings"
	"unsafe"
)

func StringToTypeValue(typeName, v string) any {
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
		return int(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int(0))))))
	case "int8":
		return int8(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int8(0))))))
	case "int16":
		return int16(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int16(0))))))
	case "int32":
		return int32(ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int32(0))))))
	case "int64":
		return ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int64(0)))))
	case "uint":
		return uint(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint(0))))))
	case "uint8":
		return uint8(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint8(0))))))
	case "uint16":
		return uint16(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint16(0))))))
	case "uint32":
		return uint32(ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint32(0))))))
	case "uint64":
		return ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "float32":
		return float32(ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float32(0))))))
	case "float64":
		return ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float64(0)))))
	case "uintptr":
		return ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "complex64":
		return complex64(ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex64(0))))))
	case "complex128":
		return ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex128(0)))))
	}
	return nil
}
