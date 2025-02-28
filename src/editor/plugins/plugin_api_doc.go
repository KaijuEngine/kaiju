//go:build editor

package plugins

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

const prefefinedAPIDocs = `---@Shape Pointer
--- Start an interactive debugger in the console window
function breakpoint() end
`

func reflectCreateDefaultForTypeName(name string) string {
	switch name {
	case "boolean":
		return "false"
	case "string":
		return `""`
	case "number":
		return "0"
	default:
		return name + ".New()"
	}
}

func reflectCommentDocCommonType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Pointer:
		//if t.Elem().Kind() == reflect.Pointer {
		return "Pointer"
		//}
		//return reflectCommentDocCommonType(t.Elem())
	default:
		return t.Name()
	}
}

func reflectCommentDocTypeHint(t reflect.Type) string {
	tName := reflectCommentDocCommonType(t)
	switch t.Kind() {
	case reflect.Map:
		// TODO:  Implement maps type
	case reflect.Array, reflect.Slice:
		if tName == "" {
			e := t.Elem()
			depth := 1
			for e.Kind() == reflect.Array || e.Kind() == reflect.Slice {
				e = e.Elem()
				depth++
			}
			tName = reflectCommentDocCommonType(t.Elem()) + strings.Repeat("[]", depth)
		}
	}
	return tName
}

func reflectStructAPI(t reflect.Type, apiOut io.StringWriter) {
	pt := reflect.PointerTo(t)
	methods := make([]reflect.Method, 0, pt.NumMethod())
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	for _, m := range methods {
		mt := m.Type
		ins := make([]string, mt.NumIn()-1)
		for i := range mt.NumIn() - 1 {
			ins[i] = fmt.Sprintf("arg%d", i)
			tName := reflectCommentDocTypeHint(mt.In(i + 1))
			apiOut.WriteString(fmt.Sprintf("---@param %s %s\n", ins[i], tName))
		}
		outs := make([]string, mt.NumOut())
		for i := range mt.NumOut() {
			o := mt.Out(i)
			tName := reflectCommentDocTypeHint(o)
			apiOut.WriteString(fmt.Sprintf("---@return %s\n", tName))
			switch o.Kind() {
			case reflect.Bool:
				outs[i] = "false"
			case reflect.String:
				outs[i] = `""`
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Float32, reflect.Float64:
				outs[i] = "0"
			case reflect.Array, reflect.Slice:
				if o.Name() != "" {
					outs[i] = o.Name() + ".New()"
				} else {
					outs[i] = "{}"
				}
			case reflect.Pointer:
				outs[i] = reflectCreateDefaultForTypeName(
					reflectCommentDocCommonType(o))
			default:
				outs[i] = o.Name() + ".New();"
			}
		}
		out := "return " + strings.Join(outs, ", ")
		apiOut.WriteString(fmt.Sprintf("function %s.%s(%s) %s end\n",
			t.Name(), m.Name, strings.Join(ins, ", "), out))
	}
	apiOut.WriteString("\n")
}

func RegenerateAPI() error {
	const apiFile = "content/editor/plugins/api.lua"
	f, err := os.OpenFile(apiFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(prefefinedAPIDocs)
	for _, t := range reflectedTypes() {
		reflectStructAPI(t, f)
	}
	return nil
}
