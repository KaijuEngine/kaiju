package details_window

import (
	"kaiju/klib"
	"log/slog"
	"strconv"
	"strings"
	"unsafe"
)

var (
	tagParsers = map[string]func(f *entityDataField, value string){
		"default": tagDefault,
		"clamp":   tagClamp,
	}
)

func tagDefault(f *entityDataField, value string) {
	f.Value = stringToValue(f.Type, value)
}

func tagClamp(f *entityDataField, value string) {
	parts := strings.Split(value, ",")
	if len(parts) == 2 {
		parts = append([]string{"0"}, parts...)
	}
	if len(parts) == 3 {
		values := make([]any, len(parts))
		for i := range parts {
			values[i] = stringToValue(f.Type, parts[i])
		}
		f.Value = values[0]
		f.Min = values[1]
		f.Max = values[2]
	} else {
		slog.Warn("invalid format for the 'clamp' tag on field", "field", f.Name)
	}
}

func stringToValue(typeName, v string) any {
	switch typeName {
	case "string":
		return v
	case "int":
		return int(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int(0))))))
	case "int8":
		return int8(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int8(0))))))
	case "int16":
		return int16(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int16(0))))))
	case "int32":
		return int32(klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int32(0))))))
	case "int64":
		return klib.ShouldReturn(strconv.ParseInt(v, 0, int(unsafe.Sizeof(int64(0)))))
	case "uint":
		return uint(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint(0))))))
	case "uint8":
		return uint8(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint8(0))))))
	case "uint16":
		return uint16(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint16(0))))))
	case "uint32":
		return uint32(klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint32(0))))))
	case "uint64":
		return klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "float32":
		return float32(klib.ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float32(0))))))
	case "float64":
		return klib.ShouldReturn(strconv.ParseFloat(v, int(unsafe.Sizeof(float64(0)))))
	case "uintptr":
		return klib.ShouldReturn(strconv.ParseUint(v, 0, int(unsafe.Sizeof(uint64(0)))))
	case "complex64":
		return complex64(klib.ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex64(0))))))
	case "complex128":
		return klib.ShouldReturn(strconv.ParseComplex(v, int(unsafe.Sizeof(complex128(0)))))
	}
	return nil
}
