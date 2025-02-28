//go:build editor

package plugins

import (
	"fmt"
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

	"github.com/KaijuEngine/go-lua"
)

type jsLogType = int

const (
	jsLogTypeDebug = jsLogType(iota)
	jsLogTypeWarn
	jsLogTypeError
)

const (
	plugins = "editor/plugins"
)

type luavm struct {
	runtime *lua.State
}

//func (vm *jsvm) throwError(message string) goja.Value {
//	if errorConstructor, ok := goja.AssertFunction(vm.runtime.Get("Error")); ok {
//		if res, err := errorConstructor(goja.Null(), vm.runtime.ToValue(message)); err == nil {
//			return res
//		}
//	}
//	return nil
//}
//
//func (vm *jsvm) logging(call goja.FunctionCall, logType jsLogType) goja.Value {
//	if len(call.Arguments) < 1 {
//		return vm.throwError("Missing arguments")
//	}
//	msg := call.Arguments[0].String()
//	prints := make([]any, len(call.Arguments[1:]))
//	for i, a := range call.Arguments[1:] {
//		prints[i] = a.Export()
//	}
//	switch logType {
//	case jsLogTypeDebug:
//		slog.Debug(msg, prints...)
//	case jsLogTypeWarn:
//		slog.Warn(msg, prints...)
//	case jsLogTypeError:
//		slog.Error(msg, prints...)
//	}
//	return nil
//}
//
//func (vm *jsvm) jsDebug(call goja.FunctionCall) goja.Value {
//	return vm.logging(call, jsLogTypeDebug)
//}
//
//func (vm *jsvm) jsWarn(call goja.FunctionCall) goja.Value {
//	return vm.logging(call, jsLogTypeWarn)
//}
//
//func (vm *jsvm) jsError(call goja.FunctionCall) goja.Value {
//	return vm.logging(call, jsLogTypeError)
//}

func reflectStructToLua[T any](vm *luavm) {
	t := reflect.TypeFor[T]()
	name := t.Name()
	vm.runtime.NewTable()
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		to := reflect.New(t)
		if state.IsTable(-1) {
			state.Field(-1, "_goPtr")
			if !state.IsUserData(1) {
				state.Pop(1)
			}
		} else if state.IsUserData(-1) {
			state.PushValue(-1)
		}
		if state.IsUserData(1) {
			ud := reflect.ValueOf(state.ToUserData(1))
			if ud.Kind() == reflect.Pointer {
				to.Elem().Set(ud.Elem())
			} else {
				to.Elem().Set(ud)
			}
			state.Pop(1)
		}
		state.Global("create_obj")
		state.Global(name)
		state.PushBoolean(false)
		state.Call(2, 1)
		if state.IsTable(-1) {
			state.PushUserData(to.Interface())
			state.SetField(-2, "_goPtr")
		}
		return 1
	})
	vm.runtime.SetField(-2, "New")
	pt := reflect.PointerTo(t)
	methods := make([]reflect.Method, 0, pt.NumMethod())
	for i := range pt.NumMethod() {
		methods = append(methods, pt.Method(i))
	}
	for _, m := range methods {
		methodName := m.Name
		mt := m.Type
		argTypes := make([]reflect.Type, mt.NumIn()-1)
		for i := range mt.NumIn() - 1 {
			argTypes[i] = mt.In(i + 1)
		}
		vm.runtime.PushGoFunction(func(state *lua.State) int {
			argCount := state.Top()
			// TODO:  Validate the inputs
			state.Field(1, "_goPtr")
			obj := state.ToUserData(-1)
			v := reflect.ValueOf(obj)
			args := make([]reflect.Value, argCount-1)
			for i := range argCount - 1 {
				idx := 2 + i
				if state.IsBoolean(idx) {
					args[i] = reflect.ValueOf(state.ToBoolean(idx))
				} else if state.IsNumber(idx) {
					n, _ := state.ToNumber(idx)
					args[i] = reflect.ValueOf(n)
					args[i] = args[i].Convert(argTypes[i])
				} else if state.IsTable(idx) {
					state.Field(idx, "_goPtr")
					args[i] = reflect.ValueOf(state.ToUserData(1))
				} else if state.IsUserData(idx) {
					args[i] = reflect.ValueOf(state.ToUserData(idx))
				} else if state.IsString(idx) {
					str, _ := state.ToString(idx)
					args[i] = reflect.ValueOf(str)
				} else {
					// TODO:  ERROR
				}
			}
			res := v.MethodByName(methodName).Call(args)
			for i := range len(res) {
				r := res[i]
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
					state.Global(rt.Name())
					state.Field(-1, "New")
					state.PushUserData(r.Interface())
					state.Call(1, 1)
				} else {
					switch kind {
					case reflect.Bool:
						state.PushBoolean(r.Interface().(bool))
					case reflect.String:
						state.PushString(r.Interface().(string))
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
						reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32,
						reflect.Uint64, reflect.Float32, reflect.Float64:
						state.PushNumber(r.Convert(reflect.TypeFor[float64]()).Interface().(float64))
					}
				}
			}
			return len(res)
		})
		vm.runtime.SetField(-2, methodName)
	}
	vm.runtime.SetGlobal(name)
}

func removeImportAPI(js string) string {
	re := regexp.MustCompile(`require\s{0,}["'][\.\/]+api(\.lua){0,}["']`)
	return re.ReplaceAllString(js, "")
}

func rollup(sandbox *os.Root, js, jsPath string, imported *[]string) (string, error) {
	// TODO:  Prevent importing already imported files
	re := regexp.MustCompile(`require\s{0,}["'](.*?)(\.lau){0,}["']`)
	matches := re.FindAllStringSubmatch(js, -1)
	imports := make([]string, 0, len(matches))
	for i := range matches {
		path := strings.TrimSpace(matches[i][1])
		if !strings.HasSuffix(path, ".lua") {
			path = path + ".lua"
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

func debugValueToString(state *lua.State) string {
	if state.IsNil(1) {
		return "nil"
	} else if state.IsBoolean(1) {
		return fmt.Sprintf("%t", state.ToBoolean(1))
	} else if state.IsNumber(1) {
		n, _ := state.ToNumber(1)
		return fmt.Sprintf("%g", n)
	} else if state.IsTable(1) {
		return "[table]" // Simplified; could expand to print table contents
	} else if state.IsFunction(1) {
		return "[function]"
	} else {
		s, _ := state.ToString(1)
		return s
	}
}

func (vm *luavm) debugHookCallback(state *lua.State, ar lua.Debug) {
	where, ok := lua.Stack(state, 0)
	if !ok {
		return
	}
	d, ok := lua.Info(state, "nSltuf", where)
	if !ok || d.CurrentLine < 0 {
		return
	}

	s := strings.Split(d.Source, "\n")
	s[d.CurrentLine] = "=> " + s[d.CurrentLine]
	fmt.Println(strings.Join(s, "\n"))
	fmt.Println("")
	lua.DoString(state, "(function() return test end)()")
	debugValueToString(state)
	var input, last string
	for {
		fmt.Print("Debug> ")
		fmt.Scanln(&input)
		if strings.TrimSpace(input) == "" {
			input = last
		}
		last = input
		switch input {
		case "p":
			state.Field(state.Top(), "v")
			debugValueToString(state)
		case "n":
			return
		case "c":
			lua.SetDebugHook(state, nil, 0, 0)
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func setupDebugEnvironment(vm *luavm) {
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		lua.SetDebugHook(state, vm.debugHookCallback,
			lua.HookCall|lua.HookLine|lua.HookReturn|lua.HookTailCall, 0)
		return 0
	})
	vm.runtime.SetGlobal("breakpoint")
}

func launchPlugin(ed interfaces.Editor, entry string) error {
	vm := &luavm{
		runtime: lua.NewState(),
	}
	lua.OpenLibraries(vm.runtime)
	reflectStructToLua[matrix.Vec2](vm)
	reflectStructToLua[matrix.Vec2i](vm)
	reflectStructToLua[matrix.Vec3](vm)
	reflectStructToLua[matrix.Vec3i](vm)
	reflectStructToLua[matrix.Vec4](vm)
	reflectStructToLua[matrix.Vec4i](vm)
	reflectStructToLua[matrix.Quaternion](vm)
	reflectStructToLua[matrix.Mat3](vm)
	reflectStructToLua[matrix.Mat4](vm)
	setupDebugEnvironment(vm)
	prereq := []string{"globals.lua"}
	for i := range prereq {
		err := lua.DoFile(vm.runtime, ed.Host().AssetDatabase().ToRawPath(
			filepath.Join(plugins, prereq[i])))
		if err != nil {
			return err
		}
	}
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
		if err := lua.DoString(vm.runtime, final); err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func LaunchPlugins(ed interfaces.Editor) error {
	pluginsPath := ed.Host().AssetDatabase().ToRawPath(plugins)
	dirs, err := os.ReadDir(pluginsPath)
	if err != nil {
		return err
	}
	for i := range dirs {
		if !dirs[i].IsDir() {
			continue
		}
		err := launchPlugin(ed, filepath.Join(pluginsPath, dirs[i].Name(), "main.lua"))
		if err != nil {
			slog.Error("plugin failed to load", "plugin", dirs[i].Name(), "error", err)
		}
	}
	return nil
}
