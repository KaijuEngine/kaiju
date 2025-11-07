/******************************************************************************/
/* ollama_reflect.go                                                          */
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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package ollama

import (
	"fmt"
	"kaiju/debug"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
)

var (
	tools  = map[string]ToolFunc{}
	enumRe = regexp.MustCompile(`\s+\[('.*?'){1,}\]`)
)

type ToolFunc struct {
	tool   Tool
	fn     any
	argIdx map[int]string
}

func ReflectFuncToOllama(fn any, name, description string, argDescPair ...string) error {
	defer tracing.NewRegion("ollama.reflectToOllama").End()
	var err error
	defer func() {
		if err != nil {
			slog.Error("there was an error registering your function", "name", name, "error", err)
		}
	}()
	fnT := reflect.ValueOf(fn).Type()
	if fnT.Kind() != reflect.Func {
		err = fmt.Errorf("the type '%s' is not a func", name)
		return err
	}
	if fnT.NumOut() == 0 || fnT.Out(0).Kind() != reflect.String {
		err = fmt.Errorf("the function expects to have a string return")
		return err
	}
	if _, ok := tools[name]; ok {
		err = fmt.Errorf("a function named '%s' has already been registered", name)
		return err
	}
	debug.Assert(len(argDescPair)/2 == fnT.NumIn(), "arg map name/description count missmatch")
	tf := ToolFunc{Tool{Type: "function"}, fn, make(map[int]string)}
	tf.tool.Function.Name = name
	tf.tool.Function.Description = description
	params := FunctionParameters{Type: "object", Properties: map[string]FunctionParameterProperty{}}
	required := []string{}
	for i := 0; i < fnT.NumIn(); i++ {
		arg := fnT.In(i)
		var jsonType string
		switch arg.Kind() {
		case reflect.String:
			jsonType = "string"
		case reflect.Bool:
			jsonType = "boolean"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			jsonType = "integer"
		case reflect.Float32, reflect.Float64:
			jsonType = "number"
		case reflect.Slice, reflect.Array:
			jsonType = "array"
		case reflect.Struct:
			jsonType = "object"
		default:
			jsonType = "string"
		}
		name := strings.TrimSpace(argDescPair[i*2])
		desc := strings.TrimSpace(argDescPair[i*2+1])
		enum := []string{}
		enumMatch := enumRe.FindAllStringSubmatch(desc, 1)
		if len(enumMatch) > 0 && len(enumMatch[0]) > 1 {
			enum = strings.Split(enumMatch[0][1], ",")
			for i := range enum {
				enum[i] = strings.Trim(strings.TrimSpace(enum[i]), "'")
			}
			desc = strings.ReplaceAll(desc, enumMatch[0][0], "")
		}
		params.Properties[name] = FunctionParameterProperty{
			Type:        jsonType,
			Description: desc,
			Enum:        enum,
		}
		tf.argIdx[i] = name
		required = append(required, name)
	}
	params.Required = required
	tf.tool.Function.Parameters = params
	tools[name] = tf
	return nil
}

func callToolFunc(call ToolCall) (string, error) {
	tool, ok := tools[call.Function.Name]
	if !ok {
		return "", fmt.Errorf("no tool named '%s' found", call.Function.Name)
	}
	fnVal := reflect.ValueOf(tool.fn)
	fnType := fnVal.Type()
	if fnType.Kind() != reflect.Func {
		return "", fmt.Errorf("stored value for '%s' is not a function", call.Function.Name)
	}
	args := make([]reflect.Value, fnType.NumIn())
	if fnType.NumIn() != len(call.Function.Arguments) {
		return "", fmt.Errorf("not enough arguments, expected %d, got %d",
			fnType.NumIn(), len(call.Function.Arguments))
	}
	for i := range fnType.NumIn() {
		paramType := fnType.In(i)
		v, ok := call.Function.Arguments[tool.argIdx[i]]
		if !ok {
			for j := range tool.argIdx {
				delete(call.Function.Arguments, tool.argIdx[j])
			}
			return "", fmt.Errorf("invalid parameter name supplied: '%s'",
				strings.Join(klib.MapKeys(call.Function.Arguments), "', '"))
		}
		val := reflect.ValueOf(v)
		if val.Type().AssignableTo(paramType) {
			args[i] = val
			continue
		}
		// Handle JSON number (float64) conversion to target numeric type
		if val.Kind() == reflect.Float64 {
			switch paramType.Kind() {
			case reflect.Int:
				args[i] = reflect.ValueOf(int(val.Float()))
			case reflect.Int8:
				args[i] = reflect.ValueOf(int8(val.Float()))
			case reflect.Int16:
				args[i] = reflect.ValueOf(int16(val.Float()))
			case reflect.Int32:
				args[i] = reflect.ValueOf(int32(val.Float()))
			case reflect.Int64:
				args[i] = reflect.ValueOf(int64(val.Float()))
			case reflect.Uint:
				args[i] = reflect.ValueOf(uint(val.Float()))
			case reflect.Uint8:
				args[i] = reflect.ValueOf(uint8(val.Float()))
			case reflect.Uint16:
				args[i] = reflect.ValueOf(uint16(val.Float()))
			case reflect.Uint32:
				args[i] = reflect.ValueOf(uint32(val.Float()))
			case reflect.Uint64:
				args[i] = reflect.ValueOf(uint64(val.Float()))
			case reflect.Float32:
				args[i] = reflect.ValueOf(float32(val.Float()))
			case reflect.Float64:
				args[i] = val
			default:
				// Unsupported conversion
				return "", fmt.Errorf("cannot convert float64 to %s for parameter %d of function '%s'", paramType.Kind(), i, call.Function.Name)
			}
			continue
		}
		// If we reach here, no suitable conversion found
		return "", fmt.Errorf("could not find suitable argument for parameter %d of function '%s'", i, call.Function.Name)
	}
	out := fnVal.Call(args)
	return out[0].String(), nil
}
