//go:build editor

package plugins

import (
	"fmt"
	"io"
	"kaiju/matrix"
	"os"
	"reflect"
	"strings"
)

const prefefinedAPIDocs = `/**
 * @param {string} msg 
 * @param  {...any} args 
 */
function debug(msg, ...args) {}
/**
 * @param {string} msg 
 * @param  {...any} args 
 */
function warn(msg, ...args) {}
/**
 * @param {string} msg 
 * @param  {...any} args 
 */
function error(msg, ...args) {}
function Pointer() {}
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
		return "new " + name + "()"
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

func reflectStructAPI[T any](apiOut io.StringWriter) {
	t := reflect.TypeFor[T]()
	apiOut.WriteString(fmt.Sprintf("function %s() {\n", t.Name()))
	pt := reflect.PointerTo(t)
	methods := make([]reflect.Method, 0, pt.NumMethod())
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	for _, m := range methods {
		mt := m.Type
		apiOut.WriteString("\t/**\n\t * @function\n")
		ins := make([]string, mt.NumIn()-1)
		out := ""
		for i := range mt.NumIn() - 1 {
			ins[i] = fmt.Sprintf("arg%d", i)
			tName := reflectCommentDocTypeHint(mt.In(i + 1))
			apiOut.WriteString(fmt.Sprintf("\t * @param {%s} %s\n", tName, ins[i]))
		}
		// TODO:  If num out is > 0 then return an array
		for i := range min(mt.NumOut(), 1) {
			o := mt.Out(i)
			tName := reflectCommentDocTypeHint(o)
			apiOut.WriteString(fmt.Sprintf("\t * @return {%s} arg%d\n", tName, i))
			switch o.Kind() {
			case reflect.Bool:
				out = "return false;"
			case reflect.String:
				out = `return "";`
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Float32, reflect.Float64:
				out = "return 0;"
			case reflect.Array, reflect.Slice:
				if o.Name() != "" {
					out = "return new " + o.Name() + "();"
				} else {
					out = "return [];"
				}
			case reflect.Pointer:
				out = "return " + reflectCreateDefaultForTypeName(
					reflectCommentDocCommonType(o)) + ";"
			default:
				out = "return new " + o.Name() + "();"
			}
		}
		apiOut.WriteString("\t */\n")
		apiOut.WriteString(fmt.Sprintf("\tthis.%s = function(%s) { %s }\n",
			m.Name, strings.Join(ins, ", "), out))
	}
	apiOut.WriteString("}\n")
}

func RegenerateAPI() error {
	const apiFile = "content/editor/plugins/api.js"
	f, err := os.OpenFile(apiFile, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(prefefinedAPIDocs)
	reflectStructAPI[matrix.Vec2](f)
	reflectStructAPI[matrix.Vec2i](f)
	reflectStructAPI[matrix.Vec3](f)
	reflectStructAPI[matrix.Vec3i](f)
	reflectStructAPI[matrix.Vec4](f)
	reflectStructAPI[matrix.Vec4i](f)
	reflectStructAPI[matrix.Quaternion](f)
	reflectStructAPI[matrix.Mat3](f)
	reflectStructAPI[matrix.Mat4](f)
	return nil
}
