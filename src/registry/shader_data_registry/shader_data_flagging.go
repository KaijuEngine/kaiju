/******************************************************************************/
/* shader_data_flagging.go                                                    */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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
