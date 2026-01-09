package shader_data_registry

import (
	"kaiju/rendering"
	"reflect"
	"unsafe"
)

type StandardShaderDataFlags = uint32

const (
	ShaderDataStandardFlagOutline = StandardShaderDataFlags(1 << iota)
	// Enable bit will be set anytime there are flags. This is needed because
	// bits at the extremes of the float will be truncated to 0 otherwise. By
	// setting this bit (largest exponent bit 2^1) this issue can be prevented.
	ShaderDataStandardFlagEnable = 1 << 30
)

func findStandardShaderDataFlags(target rendering.DrawInstance) (reflect.Value, bool) {
	if target == nil {
		return reflect.Value{}, false
	}
	val := reflect.ValueOf(target)
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return reflect.Value{}, false
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, false
	}
	flagType := reflect.TypeOf(StandardShaderDataFlags(0))
	for i := 0; i < val.NumField(); i++ {
		f := val.Type().Field(i)
		if f.Type == flagType {
			fieldVal := val.Field(i)
			if !fieldVal.CanInterface() {
				fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
			}
			return fieldVal, true
		}
	}
	return reflect.Value{}, false
}

func StandardShaderDataFlagsTest(target rendering.DrawInstance, flag StandardShaderDataFlags) bool {
	fieldVal, ok := findStandardShaderDataFlags(target)
	if !ok {
		return false
	}
	flags := fieldVal.Interface().(StandardShaderDataFlags)
	return (flags & flag) != 0
}

func StandardShaderDataFlagsSet(target rendering.DrawInstance, flag StandardShaderDataFlags) {
	fieldVal, ok := findStandardShaderDataFlags(target)
	if !ok {
		return
	}
	flags := fieldVal.Interface().(StandardShaderDataFlags)
	fieldVal.SetUint(uint64(flags | flag))
	updateStandardShaderDataFlagEnableStatus(fieldVal)
}

func StandardShaderDataFlagsClear(target rendering.DrawInstance, flag StandardShaderDataFlags) {
	fieldVal, ok := findStandardShaderDataFlags(target)
	if !ok {
		return
	}
	flags := fieldVal.Interface().(StandardShaderDataFlags)
	fieldVal.SetUint(uint64(flags &^ flag))
	updateStandardShaderDataFlagEnableStatus(fieldVal)
}

func updateStandardShaderDataFlagEnableStatus(fieldVal reflect.Value) {
	flags := fieldVal.Interface().(StandardShaderDataFlags)
	if flags|ShaderDataStandardFlagEnable == ShaderDataStandardFlagEnable {
		flags = 0
	} else {
		flags |= ShaderDataStandardFlagEnable
	}
	fieldVal.SetUint(uint64(flags))
}
