package stage_workspace

import (
	"kaiju/editor/codegen"
	"kaiju/editor/editor_stage_manager"
	"kaiju/klib"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
)

var (
	tagParsers = map[string]func(f *entityDataField, value string){
		"default": tagDefault,
		"clamp":   tagClamp,
	}
)

type entityDataEntry struct {
	gen        codegen.GeneratedType
	entityData any
	Name       string
	Fields     []entityDataField
}

type entityDataField struct {
	Idx   int
	Name  string
	Type  string
	Pkg   string
	Value any
	Min   any
	Max   any
}

func (f *entityDataField) IsNumber() bool   { return isNumber(f.Type) }
func (f *entityDataField) IsInput() bool    { return isInput(f.Type) }
func (f *entityDataField) IsCheckbox() bool { return isCheckbox(f.Type) }
func (f *entityDataField) IsEntityId() bool { return isEntityId(f.Pkg, f.Type) }

func (g *entityDataEntry) fieldNumberAsString(fieldIdx int) string {
	f := g.Fields[fieldIdx]
	if !f.IsNumber() {
		return "0"
	}
	v := g.entityData.(reflect.Value).Elem().Field(fieldIdx)
	switch f.Value.(type) {
	case int:
		return strconv.FormatInt(v.Int(), 10)
	case int8:
		return strconv.FormatInt(v.Int(), 10)
	case int16:
		return strconv.FormatInt(v.Int(), 10)
	case int32:
		return strconv.FormatInt(v.Int(), 10)
	case int64:
		return strconv.FormatInt(v.Int(), 10)
	case uint:
		return strconv.FormatUint(v.Uint(), 10)
	case uint8:
		return strconv.FormatUint(v.Uint(), 10)
	case uint16:
		return strconv.FormatUint(v.Uint(), 10)
	case uint32:
		return strconv.FormatUint(v.Uint(), 10)
	case uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case float32:
		return strconv.FormatFloat(v.Float(), 'f', 5, 32)
	case float64:
		return strconv.FormatFloat(v.Float(), 'f', 5, 64)
	}
	return "0"
}

func (g *entityDataEntry) fieldString(fieldIdx int) string {
	v := g.entityData.(reflect.Value).Elem().Field(fieldIdx)
	return v.String()
}

func (g *entityDataEntry) fieldBool(fieldIdx int) bool {
	v := g.entityData.(reflect.Value).Elem().Field(fieldIdx)
	return v.Bool()
}

func readEntityDataBindingType(g codegen.GeneratedType, e *editor_stage_manager.StageEntity) *entityDataEntry {
	v := g.New().Value
	de := &entityDataEntry{}
	for i := range g.Fields {
		if g.Fields[i].IsExported() {
			ef := entityDataField{
				Idx:   i,
				Name:  g.Fields[i].Name,
				Type:  g.Fields[i].Type.Name(),
				Pkg:   g.Fields[i].Type.PkgPath(),
				Value: v.Elem().Field(i).Interface(),
			}
			if string(g.Fields[i].Tag) != "" {
				for k, fn := range tagParsers {
					if v, ok := g.Fields[i].Tag.Lookup(k); ok {
						fn(&ef, v)
						break
					}
				}
			}
			de.Fields = append(de.Fields, ef)
			fv := v.Elem().Field(len(de.Fields) - 1)
			fv.Set(reflect.ValueOf(ef.Value))
		}
	}
	de.entityData = v
	e.AddDataBinding(de)
	return de
}

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

func toInt(str string) int64 {
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toUint(str string) uint64 {
	if str == "" {
		return 0
	}
	if i, err := strconv.ParseUint(str, 10, 64); err == nil {
		return i
	}
	return 0
}

func toFloat(str string) float64 {
	if str == "" {
		return 0
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return f
	}
	return 0
}

func isNumber(typeName string) bool {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128":
		return true
	default:
		return false
	}
}

func isInput(typeName string) bool {
	return typeName == "string" || isNumber(typeName)
}

func isCheckbox(typeName string) bool {
	return typeName == "bool"
}

func isEntityId(packageName, typeName string) bool {
	return packageName == "kaiju/engine" && typeName == "EntityId"
}
