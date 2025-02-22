package details_window

import (
	"kaiju/klib"
	"log/slog"
	"strings"
)

var (
	tagParsers = map[string]func(f *entityDataField, value string){
		"default": tagDefault,
		"clamp":   tagClamp,
	}
)

func tagDefault(f *entityDataField, value string) {
	f.Value = klib.StringToTypeValue(f.Type, value)
}

func tagClamp(f *entityDataField, value string) {
	if !f.IsNumber() {
		slog.Warn("cannot use the clamp tag on non-numeric field", "field", f.Name)
		return
	}
	parts := strings.Split(value, ",")
	if len(parts) == 2 {
		parts = append([]string{"0"}, parts...)
	}
	if len(parts) == 3 {
		values := make([]any, len(parts))
		for i := range parts {
			values[i] = klib.StringToTypeValue(f.Type, parts[i])
		}
		f.Value = values[0]
		f.Min = values[1]
		f.Max = values[2]
	} else {
		slog.Warn("invalid format for the 'clamp' tag on field", "field", f.Name)
	}
}
