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
	"kaiju/platform/profiler/tracing"
	"reflect"
)

var tools = map[string]ToolFunc{}

type ToolFunc struct {
	tool Tool
	fn   any
}

func ReflectFuncToOllama(fn any, name, description string, argDescPair ...string) error {
	fnT := reflect.ValueOf(fn).Type()
	defer tracing.NewRegion("ollama.reflectToOllama").End()
	if fnT.Kind() != reflect.Func {
		return fmt.Errorf("the type '%s' is not a func", name)
	}
	if _, ok := tools[name]; ok {
		return fmt.Errorf("a function named '%s' has already been registered", name)
	}
	debug.Assert(len(argDescPair)/2 == fnT.NumIn(), "arg map name/description count missmatch")
	tool := Tool{Type: "function"}
	tool.Function.Name = name
	tool.Function.Description = description
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
		name := argDescPair[i*2]
		desc := argDescPair[i*2+1]
		params.Properties[name] = FunctionParameterProperty{
			Type:        jsonType,
			Description: desc,
		}
		required = append(required, name)
	}
	params.Required = required
	tool.Function.Parameters = params
	tools[name] = ToolFunc{tool, fn}
	return nil
}
