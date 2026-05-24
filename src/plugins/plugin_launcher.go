/******************************************************************************/
/* plugin_launcher.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package plugins

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/plugins/lua"
)

type jsLogType = int

const (
	jsLogTypeDebug = jsLogType(iota)
	jsLogTypeWarn
	jsLogTypeError
)

const (
	plugins            = "plugins"
	globalCleanupPtrFn = "__kaiju_engine_cleanup_go_ptr__"
	goPtrField         = "_goPtr"
)

var (
	apiReg     = regexp.MustCompile(`require\s{0,}\({0,1}\s{0,}["'][\.\/]+api(\.lua){0,}["']\s{0,}\){0,1}`)
	requireReg = regexp.MustCompile(`require\s{0,}\({0,1}\s{0,}["'](.*?)(\.lua){0,}["']\s{0,}\){0,1}`)
)

type LuaVM struct {
	PluginPath string
	runtime    lua.State
	sandbox    *os.Root
}

func (vm *LuaVM) InvokeGlobalFunction(name string) {
	_ = vm.InvokeGlobalFunctionWithArgs(name)
}

func (vm *LuaVM) InvokeGlobalFunctionWithArgs(name string, args ...reflect.Value) error {
	vm.runtime.Global(name)
	if vm.runtime.IsFunction(-1) {
		for _, arg := range args {
			vm.pushReflectValue(arg)
		}
		return vm.runtime.ProtectedCall(len(args), 0)
	} else {
		vm.runtime.Pop(1)
	}
	return nil
}

func (vm *LuaVM) Close() {
	if vm.sandbox != nil {
		vm.sandbox.Close()
		vm.sandbox = nil
	}
	vm.runtime.Close()
}

func (vm *LuaVM) pushReflectValue(v reflect.Value) {
	if !v.IsValid() {
		vm.runtime.PushBoolean(false)
		return
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Bool:
		vm.runtime.PushBoolean(v.Bool())
	case reflect.String:
		vm.runtime.PushString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vm.runtime.PushNumber(float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vm.runtime.PushNumber(float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		vm.runtime.PushNumber(v.Convert(reflect.TypeFor[float64]()).Float())
	default:
		wrapReflectValueForLua(v, &vm.runtime)
	}
}

func unwrapLuaTablePointer(state *lua.State, idx int) (any, bool) {
	state.Field(idx, goPtrField)
	defer state.Pop(1)
	if !state.IsUserData(-1) {
		return nil, false
	}
	return state.ToUserData(-1), true
}

func luaValueToReflect(state *lua.State, idx int, target reflect.Type) (reflect.Value, error) {
	if state.IsTable(idx) {
		obj, ok := unwrapLuaTablePointer(state, idx)
		if !ok {
			return reflect.Value{}, fmt.Errorf("argument %d is a table without %s", idx, goPtrField)
		}
		v := reflect.ValueOf(obj)
		if v.Type().AssignableTo(target) {
			return v, nil
		}
		if v.Kind() == reflect.Pointer && v.Elem().Type().AssignableTo(target) {
			return v.Elem(), nil
		}
		return reflect.Value{}, fmt.Errorf("argument %d is %s, expected %s", idx, v.Type(), target)
	}
	if state.IsUserData(idx) {
		v := reflect.ValueOf(state.ToUserData(idx))
		if v.Type().AssignableTo(target) {
			return v, nil
		}
		if v.Kind() == reflect.Pointer && v.Elem().Type().AssignableTo(target) {
			return v.Elem(), nil
		}
		return reflect.Value{}, fmt.Errorf("argument %d is %s, expected %s", idx, v.Type(), target)
	}
	if state.IsBoolean(idx) {
		if target.Kind() != reflect.Bool {
			return reflect.Value{}, fmt.Errorf("argument %d is boolean, expected %s", idx, target)
		}
		return reflect.ValueOf(state.ToBoolean(idx)).Convert(target), nil
	}
	if state.IsString(idx) && target.Kind() == reflect.String {
		return reflect.ValueOf(state.ToString(idx)).Convert(target), nil
	}
	if state.IsNumber(idx) {
		n := state.ToNumber(idx)
		v := reflect.New(target).Elem()
		switch target.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(int64(n))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if n < 0 {
				return reflect.Value{}, fmt.Errorf("argument %d is negative, expected %s", idx, target)
			}
			v.SetUint(uint64(n))
		case reflect.Float32, reflect.Float64:
			v.SetFloat(n)
		default:
			return reflect.Value{}, fmt.Errorf("argument %d is number, expected %s", idx, target)
		}
		return v, nil
	}
	return reflect.Value{}, fmt.Errorf("argument %d has unsupported Lua type for %s", idx, target)
}

func wrapReflectValueForLua(v reflect.Value, state *lua.State) int {
	if !v.IsValid() {
		state.PushBoolean(false)
		return 1
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	rt := v.Type()
	switch rt.Kind() {
	case reflect.Bool:
		state.PushBoolean(v.Bool())
	case reflect.String:
		state.PushString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		state.PushNumber(float64(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		state.PushNumber(float64(v.Uint()))
	case reflect.Float32, reflect.Float64:
		state.PushNumber(v.Convert(reflect.TypeFor[float64]()).Float())
	case reflect.Pointer:
		if v.IsNil() {
			state.PushBoolean(false)
			return 1
		}
		state.Global(rt.Elem().Name())
		state.Field(-1, "New")
		state.PushUserData(v)
		if err := state.ProtectedCall(1, 1); err != nil {
			return state.Error(err.Error())
		}
		state.Remove(-2)
	case reflect.Array, reflect.Struct:
		if rt.Kind() == reflect.Array && rt.Name() == "" {
			state.CreateTable(v.Len(), 0)
			for i := 0; i < v.Len(); i++ {
				wrapReflectValueForLua(v.Index(i), state)
				state.RawSetInt(-2, i+1)
			}
			return 1
		}
		rp := reflect.New(rt)
		rp.Elem().Set(v)
		state.Global(rt.Name())
		state.Field(-1, "New")
		state.PushUserData(rp)
		if err := state.ProtectedCall(1, 1); err != nil {
			return state.Error(err.Error())
		}
		state.Remove(-2)
	case reflect.Slice:
		state.CreateTable(v.Len(), 0)
		for i := 0; i < v.Len(); i++ {
			wrapReflectValueForLua(v.Index(i), state)
			state.RawSetInt(-2, i+1)
		}
	default:
		state.PushString(fmt.Sprintf("<unsupported %s>", rt))
	}
	return 1
}

func reflectStructToLua(t reflect.Type, vm *LuaVM) {
	defer tracing.NewRegion("plugins.reflectStructToLua").End()
	name := t.Name()
	vm.runtime.NewTable()
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		var to reflect.Value
		if state.IsUserData(-1) {
			to = reflect.ValueOf(state.ToUserData(-1))
		} else {
			to = reflect.New(t)
			if t.Kind() == reflect.Array {
				for i := 1; i <= state.Top() && i <= t.Len(); i++ {
					v, err := luaValueToReflect(state, i, t.Elem())
					if err != nil {
						return state.Error(err.Error())
					}
					to.Elem().Index(i - 1).Set(v)
				}
			}
		}
		state.Global("create_obj")
		state.Global(name)
		state.PushBoolean(false)
		if err := state.ProtectedCall(2, 1); err != nil {
			return state.Error(err.Error())
		}
		if state.IsTable(-1) {
			state.PushUserData(to)
			state.SetField(-2, goPtrField)
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
			if argCount-1 != len(argTypes) {
				return state.Error(fmt.Sprintf("%s.%s expects %d arguments, got %d",
					name, methodName, len(argTypes), argCount-1))
			}
			state.Field(1, goPtrField)
			if !state.IsUserData(-1) {
				return state.Error(fmt.Sprintf("%s.%s called without a Go object", name, methodName))
			}
			obj := state.ToUserData(-1)
			state.Pop(1)
			v := reflect.ValueOf(obj)
			args := make([]reflect.Value, argCount-1)
			for i := range argCount - 1 {
				idx := 2 + i
				arg, err := luaValueToReflect(state, idx, argTypes[i])
				if err != nil {
					return state.Error(fmt.Sprintf("%s.%s: %s", name, methodName, err.Error()))
				}
				args[i] = arg
			}
			res := v.MethodByName(methodName).Call(args)
			for i := range len(res) {
				wrapReflectValueForLua(res[i], state)
			}
			return len(res)
		})
		vm.runtime.SetField(-2, methodName)
	}
	vm.runtime.Global(globalCleanupPtrFn)
	vm.runtime.SetField(-2, "__gc")
	vm.runtime.SetGlobal(name)
}

func removeImportAPI(lua string) string {
	defer tracing.NewRegion("plugins.removeImportAPI").End()
	return apiReg.ReplaceAllString(lua, "")
}

func (vm *LuaVM) rollup(lua, luaPath string, imported *[]string) error {
	defer tracing.NewRegion("LuaVM.rollup").End()
	lua = removeImportAPI(lua)
	matches := requireReg.FindAllStringSubmatch(lua, -1)
	imports := make([]string, 0, len(matches))
	for i := range matches {
		path := strings.TrimSpace(matches[i][1])
		if !strings.HasSuffix(path, ".lua") {
			path = path + ".lua"
		}
		imports = append(imports, path)
	}
	for i := range imports {
		fullPath := filepath.Join(luaPath, imports[i])
		if slices.Contains(*imported, fullPath) {
			continue
		}
		path := strings.TrimPrefix(filepath.ToSlash(
			strings.TrimPrefix(fullPath, vm.sandbox.Name())), "/")
		importFile, err := vm.sandbox.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to import (%s) on file (%s)", imports[i], luaPath)
		}
		defer importFile.Close()
		inner, err := io.ReadAll(importFile)
		if err != nil {
			return fmt.Errorf("failed to read import (%s) for file (%s)", imports[i], luaPath)
		}
		innerLua := string(inner)
		if err := vm.rollup(innerLua, filepath.Dir(fullPath), imported); err != nil {
			slog.Error("error importing", "error", err)
			continue
		}
		*imported = append(*imported, fullPath)
	}
	return nil
}

func (vm *LuaVM) setupPrerequisites(adb assets.Database) error {
	defer tracing.NewRegion("LuaVM.setupPrerequisites").End()
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		if state.IsTable(-1) {
			state.Field(-1, goPtrField)
			if state.IsUserData(-1) {
				vm.runtime.RemovePinnedPointer(-1)
			}
		}
		return 0
	})
	vm.runtime.SetGlobal(globalCleanupPtrFn)
	prereq := []string{"globals.lua"}
	for i := range prereq {
		s, err := adb.ReadText(filepath.Join(plugins, prereq[i]))
		if err != nil {
			return err
		}
		if err = vm.runtime.DoString(s); err != nil {
			return err
		}
	}
	return nil
}

func newLuaVM(adb assets.Database, exposed []reflect.Type) (*LuaVM, error) {
	vm := &LuaVM{
		runtime: lua.New(),
	}
	if err := vm.runtime.OpenLibraries(); err != nil {
		return vm, err
	}
	if err := vm.setupPrerequisites(adb); err != nil {
		return vm, err
	}
	for _, t := range append(reflectedTypes(), exposed...) {
		reflectStructToLua(t, vm)
	}
	return vm, nil
}

func LaunchScript(adb assets.Database, entry string, exposed []reflect.Type, globals map[string]reflect.Value) (*LuaVM, error) {
	vm, err := newLuaVM(adb, exposed)
	if err != nil {
		return vm, err
	}
	vm.PluginPath = entry
	root := filepath.Dir(entry)
	sandbox, err := os.OpenRoot(root)
	if err != nil {
		return vm, err
	}
	vm.sandbox = sandbox
	for name, value := range globals {
		vm.pushReflectValue(value)
		vm.runtime.SetGlobal(name)
	}
	luaBytes, err := os.ReadFile(entry)
	if err != nil {
		return vm, err
	}
	imports := []string{}
	if err := vm.rollup(string(luaBytes), root, &imports); err != nil {
		return vm, err
	}
	imports = append(imports, entry)
	for i := range imports {
		if err := vm.runtime.DoFile(filepath.ToSlash(imports[i])); err != nil {
			return vm, err
		}
	}
	return vm, nil
}

func launchPlugin(adb assets.Database, entry string) (*LuaVM, error) {
	defer tracing.NewRegion("plugins.launchPlugin").End()
	vm, err := newLuaVM(adb, nil)
	vm.PluginPath = entry
	if err != nil {
		return vm, err
	}
	if lua, err := os.ReadFile(entry); err == nil {
		root := filepath.Dir(entry)
		sandbox, err := os.OpenRoot(root)
		if err != nil {
			return vm, err
		}
		vm.sandbox = sandbox
		mainLua := string(lua)
		imports := []string{}
		if err := vm.rollup(mainLua, root, &imports); err != nil {
			return vm, err
		}
		imports = append(imports, entry)
		for i := range imports {
			if err := vm.runtime.DoFile(filepath.ToSlash(imports[i])); err != nil {
				return vm, err
			}
		}
	} else {
		return vm, err
	}
	return vm, nil
}

func LaunchPlugins(adb assets.Database, path string) ([]*LuaVM, error) {
	defer tracing.NewRegion("plugins.LaunchPlugins").End()
	dirs, err := os.ReadDir(path)
	vms := make([]*LuaVM, 0)
	if err != nil {
		return vms, err
	}
	for i := range dirs {
		if !dirs[i].IsDir() {
			continue
		}
		vm, err := launchPlugin(adb, filepath.Join(path, dirs[i].Name(), "main.lua"))
		vms = append(vms, vm)
		if err != nil {
			slog.Error("plugin failed to load", "plugin", dirs[i].Name(), "error", err)
		}
	}
	return vms, nil
}
