//go:build editor

package plugins

import (
	"io"
	"kaiju/editor/interfaces"
	"kaiju/matrix"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/dop251/goja"
)

type jsLogType = int

const (
	jsLogTypeDebug = jsLogType(iota)
	jsLogTypeWarn
	jsLogTypeError
)

type jsvm struct {
	runtime *goja.Runtime
}

func (vm *jsvm) throwError(message string) goja.Value {
	if errorConstructor, ok := goja.AssertFunction(vm.runtime.Get("Error")); ok {
		if res, err := errorConstructor(goja.Null(), vm.runtime.ToValue(message)); err == nil {
			return res
		}
	}
	return nil
}

func (vm *jsvm) logging(call goja.FunctionCall, logType jsLogType) goja.Value {
	if len(call.Arguments) < 1 {
		return vm.throwError("Missing arguments")
	}
	msg := call.Arguments[0].String()
	prints := make([]any, len(call.Arguments[1:]))
	for i, a := range call.Arguments[1:] {
		prints[i] = a.Export()
	}
	switch logType {
	case jsLogTypeDebug:
		slog.Debug(msg, prints...)
	case jsLogTypeWarn:
		slog.Warn(msg, prints...)
	case jsLogTypeError:
		slog.Error(msg, prints...)
	}
	return nil
}

func (vm *jsvm) jsDebug(call goja.FunctionCall) goja.Value {
	return vm.logging(call, jsLogTypeDebug)
}

func (vm *jsvm) jsWarn(call goja.FunctionCall) goja.Value {
	return vm.logging(call, jsLogTypeWarn)
}

func (vm *jsvm) jsError(call goja.FunctionCall) goja.Value {
	return vm.logging(call, jsLogTypeError)
}

func reflectStructToJS[T any](vm *jsvm) {
	t := reflect.TypeFor[T]()
	name := t.Name()
	proto := vm.runtime.NewObject()
	constructorFunc := vm.runtime.ToValue(func(call goja.ConstructorCall) *goja.Object {
		to := reflect.New(t)
		if len(call.Arguments) > 0 {
			from := reflect.ValueOf(call.Arguments[0].Export())
			if to.Kind() == from.Kind() {
				to.Elem().Set(from)
			} else if from.Kind() != reflect.Pointer {
				to.Elem().Set(from)
			}
		}
		call.This.Set("_goPtr", to.Interface())
		return call.This
	})
	constructorObj := constructorFunc.ToObject(vm.runtime)
	constructorObj.Set("prototype", proto)
	vm.runtime.Set(name, constructorFunc)
	pt := reflect.PointerTo(t)
	methods := make([]reflect.Method, 0, pt.NumMethod())
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	for _, m := range methods {
		methodName := m.Name
		mt := m.Type
		argTypes := make([]reflect.Type, mt.NumIn()-1)
		for j := range mt.NumIn() - 1 {
			argTypes[j] = mt.In(j + 1)
		}
		proto.Set(methodName, func(call goja.FunctionCall) goja.Value {
			obj := call.This.ToObject(vm.runtime).Get("_goPtr").Export()
			v := reflect.ValueOf(obj)
			args := make([]reflect.Value, len(call.Arguments))
			for j := range call.Arguments {
				args[j] = reflect.ValueOf(call.Arguments[j].Export())
				args[j] = args[j].Convert(argTypes[j])
			}
			res := v.MethodByName(methodName).Call(args)
			if len(res) > 1 {
				slog.Error("reflected go method to JS contains too many returns", "count", len(res), "method", methodName)
			} else if len(res) == 0 {
				return nil
			}
			r := res[0]
			rt := r.Type()
			mCount := rt.NumMethod()
			kind := rt.Kind()
			switch kind {
			case reflect.Array:
			case reflect.Pointer:
				mCount += rt.Elem().NumMethod()
			default:
				mCount += reflect.PointerTo(rt).NumMethod()
			}
			if mCount > 0 {
				s := vm.runtime.Get(rt.Name())
				out, _ := vm.runtime.New(s, vm.runtime.ToValue(r.Interface()))
				return out
			} else {
				return vm.runtime.ToValue(r.Interface())
			}
		})
	}
}

func removeImportAPI(js string) string {
	re := regexp.MustCompile(`import\s{0,}["'][\.\/]+api(\.js){0,}["']`)
	return re.ReplaceAllString(js, "")
}

func rollup(sandbox *os.Root, js, jsPath string, imported *[]string) (string, error) {
	// TODO:  Prevent importing already imported files
	re := regexp.MustCompile(`import\s{0,}["'](.*?)(\.js){0,}["']`)
	matches := re.FindAllStringSubmatch(js, -1)
	imports := make([]string, 0, len(matches))
	for i := range matches {
		path := strings.TrimSpace(matches[i][1])
		if !strings.HasSuffix(path, ".js") {
			path = path + ".js"
		}
		imports = append(imports, path)
	}
	for i := range imports {
		fullPath := filepath.Join(jsPath, imports[i])
		if slices.Contains(*imported, fullPath) {
			continue
		}
		*imported = append(*imported, fullPath)
		path := strings.TrimPrefix(filepath.ToSlash(
			strings.TrimPrefix(fullPath, sandbox.Name())), "/")
		importFile, err := sandbox.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			slog.Error("failed to import file", "file", jsPath, "import", imports[i])
			continue
		}
		defer importFile.Close()
		inner, err := io.ReadAll(importFile)
		if err != nil {
			slog.Error("failed to read import file", "file", jsPath, "import", imports[i])
			continue
		}
		innerJS := removeImportAPI(string(inner))
		innerJS, _ = rollup(sandbox, innerJS, filepath.Dir(fullPath), imported)
		js = strings.ReplaceAll(js, matches[i][0], innerJS)
	}
	return re.ReplaceAllString(js, ""), nil
}

func launchPlugin(ed interfaces.Editor, entry string) error {
	vm := &jsvm{
		runtime: goja.New(),
	}
	reflectStructToJS[matrix.Vec2](vm)
	reflectStructToJS[matrix.Vec2i](vm)
	reflectStructToJS[matrix.Vec3](vm)
	reflectStructToJS[matrix.Vec3i](vm)
	reflectStructToJS[matrix.Vec4](vm)
	reflectStructToJS[matrix.Vec4i](vm)
	reflectStructToJS[matrix.Quaternion](vm)
	reflectStructToJS[matrix.Mat3](vm)
	reflectStructToJS[matrix.Mat4](vm)
	vm.runtime.Set("debug", vm.jsDebug)
	vm.runtime.Set("warn", vm.jsWarn)
	vm.runtime.Set("error", vm.jsError)
	if js, err := os.ReadFile(entry); err == nil {
		root := filepath.Dir(entry)
		sandbox, err := os.OpenRoot(root)
		if err != nil {
			return err
		}
		final, err := rollup(sandbox, removeImportAPI(string(js)), root, &[]string{})
		if err != nil {
			return err
		}
		if _, err := vm.runtime.RunString(final); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func LaunchPlugins(ed interfaces.Editor) error {
	const plugins = "editor/plugins"
	pluginsPath := ed.Host().AssetDatabase().ToRawPath(plugins)
	dirs, err := os.ReadDir(pluginsPath)
	if err != nil {
		return err
	}
	for i := range dirs {
		if !dirs[i].IsDir() {
			continue
		}
		err := launchPlugin(ed, filepath.Join(pluginsPath, dirs[i].Name(), "main.js"))
		if err != nil {
			slog.Error("plugin failed to load", "plugin", dirs[i].Name(), "error", err)
		}
	}
	return nil
}
