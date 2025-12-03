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
	"encoding/json"
	"fmt"
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
	action LLMAction
	argIdx map[int]string
}

type LLMAction interface {
	Execute() (any, error)
}

type LLMActionResultBase struct {
	Success bool   `json:"success" llmskip:"true"`
	Error   string `json:"error" llmskip:"true"`
}

type FieldInfo struct {
	Type     reflect.Type
	JSON     string
	Desc     string
	Enum     []string
	Optional bool
}

func (a *LLMActionResultBase) SetSuccess()                { a.Success = true }
func (a *LLMActionResultBase) SetErrorMessage(msg string) { a.Error = msg }

func extractJSONDescTags(a any) ([]FieldInfo, error) {
	t := reflect.TypeOf(a)
	if t == nil {
		return nil, fmt.Errorf("nil value provided")
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or pointer to struct, got %s", t.Kind())
	}
	var out []FieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if _, ok := f.Tag.Lookup("llmskip"); ok {
			continue
		}
		jsonTag := f.Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		jsonTag = strings.TrimSpace(jsonTag)
		descTag := f.Tag.Get("desc")
		descTag = strings.TrimSpace(descTag)
		out = append(out, FieldInfo{
			Type:     f.Type,
			JSON:     jsonTag,
			Desc:     descTag,
			Enum:     strings.Split(f.Tag.Get("enum"), ","),
			Optional: f.Tag.Get("optional") != "",
		})
	}
	return out, nil
}

func ReflectFuncToOllama(a LLMAction, name, description string) error {
	defer tracing.NewRegion("ollama.reflectToOllama").End()
	var err error
	defer func() {
		if err != nil {
			slog.Error("there was an error registering your function", "name", name, "error", err)
		}
	}()
	if _, ok := tools[name]; ok {
		err = fmt.Errorf("a function named '%s' has already been registered", name)
		return err
	}
	info, err := extractJSONDescTags(a)
	if err != nil {
		return err
	}
	var act LLMAction
	if reflect.TypeOf(a).Kind() == reflect.Pointer {
		act = a
	} else {
		origVal := reflect.ValueOf(a)
		copyPtr := reflect.New(origVal.Type())
		copyPtr.Elem().Set(origVal)
		act = copyPtr.Interface().(LLMAction)
	}
	tf := ToolFunc{
		tool: Tool{
			Type: "function",
			Function: Function{
				Name:        name,
				Description: description,
			},
		},
		action: act,
		argIdx: make(map[int]string),
	}
	params := FunctionParameters{
		Type:       "object",
		Properties: map[string]FunctionParameterProperty{},
	}
	required := []string{}
	for i := range info {
		var jsonType string
		switch info[i].Type.Kind() {
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
		name := info[i].JSON
		prop := FunctionParameterProperty{
			Type:        jsonType,
			Description: info[i].Desc,
			Enum:        info[i].Enum,
		}
		if len(info[i].Enum) > 0 {
			prop.Description += fmt.Sprintf("(enum ['%s'])", strings.Join(info[i].Enum, "','"))
		}
		params.Properties[name] = prop
		tf.argIdx[i] = name
		if !info[i].Optional {
			required = append(required, name)
		}
	}
	params.Required = required
	tf.tool.Function.Parameters = params
	tools[name] = tf
	return nil
}

func callToolFunc(call ToolCall) (any, error) {
	tool, ok := tools[call.Function.Name]
	if !ok {
		return nil, fmt.Errorf("no tool named '%s' found", call.Function.Name)
	}
	j, err := json.Marshal(call.Function.Arguments)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(j, tool.action); err != nil {
		return nil, err
	}
	return tool.action.Execute()
}
