/******************************************************************************/
/* plugin_launcher.go                                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package plugins

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
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

type LuaVM struct {
	PluginPath string
	runtime    lua.State
	sandbox    *os.Root
}

func (vm *LuaVM) Close() {
	if vm == nil {
		return
	}
	vm.runtime.Close()
	if vm.sandbox != nil {
		vm.sandbox.Close()
		vm.sandbox = nil
	}
}

func (vm *LuaVM) InvokeGlobalFunction(name string) {
	// TODO:  Support arguments...
	vm.runtime.Global(name)
	if vm.runtime.IsFunction(-1) {
		if err := vm.runtime.Call(0, 0); err != nil {
			slog.Error("failed to invoke lua function", "function", name, "error", err)
		}
	} else {
		vm.runtime.Pop(1)
	}
}

func (vm *LuaVM) DoStringNamed(code, name string) error {
	return vm.runtime.DoStringNamed(code, name)
}

func (vm *LuaVM) SetGlobalGoFunction(name string, fn func(*lua.State) int) {
	vm.runtime.PushGoFunction(fn)
	vm.runtime.SetGlobal(name)
}

func NewScriptVM(adb assets.Database, root string) (*LuaVM, error) {
	defer tracing.NewRegion("plugins.NewScriptVM").End()
	vm := &LuaVM{
		PluginPath: root,
		runtime:    lua.New(),
	}
	if err := vm.runtime.OpenLibraries(); err != nil {
		return vm, err
	}
	if root != "" {
		sandbox, err := os.OpenRoot(root)
		if err != nil {
			return vm, err
		}
		vm.sandbox = sandbox
	}
	if err := vm.setupPrerequisites(adb); err != nil {
		return vm, err
	}
	if vm.sandbox != nil {
		if err := vm.setupRequire(); err != nil {
			return vm, err
		}
	}
	vm.runtime.SandboxLibraries()
	for _, t := range reflectedTypes() {
		reflectStructToLua(t, vm)
	}
	return vm, nil
}

func reflectStructToLua(t reflect.Type, vm *LuaVM) {
	defer tracing.NewRegion("plugins.reflectStructToLua").End()
	name := t.Name()
	vm.runtime.NewTable()
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		to := reflect.Value{}
		if state.Top() == 1 && state.IsUserData(1) {
			ptr := state.ToUserData(1)
			if ptr == nil {
				return state.ArgError(1, "expected pinned Go pointer")
			}
			to = reflect.ValueOf(ptr)
		} else {
			to = reflect.New(t)
			if errArg, err := fillConstructorArgs(state, to.Elem(), 1, state.Top()); err != nil {
				return state.ArgError(errArg, err.Error())
			}
		}
		state.Global("create_obj")
		if !state.IsFunction(-1) {
			return state.Error("missing create_obj prerequisite")
		}
		state.Global(name)
		state.PushBoolean(false)
		if err := state.Call(2, 1); err != nil {
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
				return state.ArgError(argCount, fmt.Sprintf(
					"%s:%s expects %d arguments, got %d",
					name, methodName, len(argTypes), argCount-1))
			}
			if !state.IsTable(1) {
				return state.ArgError(1, "expected reflected object table")
			}
			state.Field(1, goPtrField)
			if !state.IsUserData(-1) {
				state.Pop(1)
				return state.ArgError(1, "missing reflected Go pointer")
			}
			obj := state.ToUserData(-1)
			state.Pop(1)
			if obj == nil {
				return state.ArgError(1, "invalid reflected Go pointer")
			}
			v := reflect.ValueOf(obj)
			args := make([]reflect.Value, argCount-1)
			for i := range argCount - 1 {
				idx := 2 + i
				arg, err := luaValueToReflect(state, idx, argTypes[i])
				if err != nil {
					return state.ArgError(idx, fmt.Sprintf(
						"%s:%s argument %d: %s",
						name, methodName, i+1, err.Error()))
				}
				args[i] = arg
			}
			method := v.MethodByName(methodName)
			if !method.IsValid() {
				return state.Error(fmt.Sprintf("method %s:%s not found", name, methodName))
			}
			res := method.Call(args)
			for i := range len(res) {
				if err := pushReflectValue(state, res[i]); err != nil {
					return state.Error(fmt.Sprintf(
						"%s:%s return %d: %s",
						name, methodName, i+1, err.Error()))
				}
			}
			return len(res)
		})
		vm.runtime.SetField(-2, methodName)
	}
	vm.runtime.Global(globalCleanupPtrFn)
	vm.runtime.SetField(-2, "__gc")
	vm.runtime.SetGlobal(name)
}

func hasLuaWrapper(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		return t.NumMethod() > 0 || t.Elem().NumMethod() > 0
	}
	if t.NumMethod() > 0 {
		return true
	}
	return reflect.PointerTo(t).NumMethod() > 0
}

func wrapperTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.Name()
}

func pushReflectValue(state *lua.State, v reflect.Value) error {
	if !v.IsValid() {
		state.PushNil()
		return nil
	}
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			state.PushNil()
			return nil
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			state.PushNil()
			return nil
		}
		if hasLuaWrapper(v.Type()) {
			state.Global(wrapperTypeName(v.Type()))
			state.Field(-1, "New")
			state.PushUserData(v)
			if err := state.Call(1, 1); err != nil {
				return err
			}
			state.Remove(-2)
			return nil
		}
	}
	if hasLuaWrapper(v.Type()) {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		state.Global(wrapperTypeName(v.Type()))
		state.Field(-1, "New")
		state.PushUserData(ptr)
		if err := state.Call(1, 1); err != nil {
			return err
		}
		state.Remove(-2)
		return nil
	}
	switch v.Kind() {
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
	case reflect.Array, reflect.Slice:
		state.CreateTable(v.Len(), 0)
		for i := range v.Len() {
			if err := pushReflectValue(state, v.Index(i)); err != nil {
				return err
			}
			state.RawSetI(-2, i+1)
		}
	default:
		return fmt.Errorf("unsupported return type %s", v.Type())
	}
	return nil
}

func fillConstructorArgs(state *lua.State, dst reflect.Value, start, end int) (int, error) {
	if start > end {
		return 0, nil
	}
	if dst.Kind() == reflect.Array || dst.Kind() == reflect.Slice {
		if start == end && state.IsTable(start) {
			return fillArrayFromLuaTable(state, start, dst)
		}
		count := end - start + 1
		if count > dst.Len() {
			return end, fmt.Errorf("expected at most %d values", dst.Len())
		}
		for i := range count {
			v, err := luaValueToReflect(state, start+i, dst.Type().Elem())
			if err != nil {
				return start + i, err
			}
			dst.Index(i).Set(v)
		}
		return 0, nil
	}
	if start <= end {
		return start, fmt.Errorf("%s does not accept constructor arguments", dst.Type())
	}
	return 0, nil
}

func fillArrayFromLuaTable(state *lua.State, idx int, dst reflect.Value) (int, error) {
	idx = state.AbsIndex(idx)
	names := []string{"x", "y", "z", "w"}
	for i := range dst.Len() {
		state.RawGetI(idx, i+1)
		if state.IsNil(-1) && i < len(names) {
			state.Pop(1)
			state.Field(idx, names[i])
		}
		v, err := luaValueToReflect(state, -1, dst.Type().Elem())
		state.Pop(1)
		if err != nil {
			return idx, fmt.Errorf("expected table element %d: %w", i+1, err)
		}
		dst.Index(i).Set(v)
	}
	return 0, nil
}

func luaGoPointer(state *lua.State, idx int) (reflect.Value, bool) {
	idx = state.AbsIndex(idx)
	if state.IsUserData(idx) {
		ptr := state.ToUserData(idx)
		if ptr == nil {
			return reflect.Value{}, false
		}
		return reflect.ValueOf(ptr), true
	}
	if state.IsTable(idx) {
		state.Field(idx, goPtrField)
		defer state.Pop(1)
		if state.IsUserData(-1) {
			ptr := state.ToUserData(-1)
			if ptr == nil {
				return reflect.Value{}, false
			}
			return reflect.ValueOf(ptr), true
		}
	}
	return reflect.Value{}, false
}

func luaValueToReflect(state *lua.State, idx int, target reflect.Type) (reflect.Value, error) {
	if ptr, ok := luaGoPointer(state, idx); ok {
		if ptr.Type().AssignableTo(target) {
			return ptr, nil
		}
		if ptr.Type().ConvertibleTo(target) {
			return ptr.Convert(target), nil
		}
		if target.Kind() != reflect.Pointer && ptr.Kind() == reflect.Pointer &&
			ptr.Elem().Type().AssignableTo(target) {
			return ptr.Elem(), nil
		}
		return reflect.Value{}, fmt.Errorf("expected %s, got %s", target, ptr.Type())
	}
	if state.IsNil(idx) {
		if target.Kind() == reflect.Pointer || target.Kind() == reflect.Interface ||
			target.Kind() == reflect.Slice || target.Kind() == reflect.Map {
			return reflect.Zero(target), nil
		}
		return reflect.Value{}, fmt.Errorf("expected %s, got nil", target)
	}
	switch target.Kind() {
	case reflect.Bool:
		if !state.IsBoolean(idx) {
			return reflect.Value{}, fmt.Errorf("expected boolean")
		}
		return reflect.ValueOf(state.ToBoolean(idx)).Convert(target), nil
	case reflect.String:
		if !state.IsString(idx) {
			return reflect.Value{}, fmt.Errorf("expected string")
		}
		return reflect.ValueOf(state.ToString(idx)).Convert(target), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if !state.IsNumber(idx) {
			return reflect.Value{}, fmt.Errorf("expected number")
		}
		v := reflect.New(target).Elem()
		v.SetInt(int64(state.ToNumber(idx)))
		return v, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if !state.IsNumber(idx) {
			return reflect.Value{}, fmt.Errorf("expected number")
		}
		v := reflect.New(target).Elem()
		v.SetUint(uint64(state.ToNumber(idx)))
		return v, nil
	case reflect.Float32, reflect.Float64:
		if !state.IsNumber(idx) {
			return reflect.Value{}, fmt.Errorf("expected number")
		}
		v := reflect.New(target).Elem()
		v.SetFloat(state.ToNumber(idx))
		return v, nil
	case reflect.Array:
		if !state.IsTable(idx) {
			return reflect.Value{}, fmt.Errorf("expected table")
		}
		v := reflect.New(target).Elem()
		_, err := fillArrayFromLuaTable(state, idx, v)
		return v, err
	case reflect.Slice:
		if !state.IsTable(idx) {
			return reflect.Value{}, fmt.Errorf("expected table")
		}
		return reflect.Value{}, fmt.Errorf("slice arguments are not supported yet")
	default:
		return reflect.Value{}, fmt.Errorf("unsupported argument type %s", target)
	}
}

func (vm *LuaVM) setupPrerequisites(adb assets.Database) error {
	defer tracing.NewRegion("LuaVM.setupPrerequisites").End()
	vm.runtime.PushGoFunction(func(state *lua.State) int { return 0 })
	vm.runtime.SetGlobal(globalCleanupPtrFn)
	prereq := []string{"debugger.lua", "globals.lua"}
	for i := range prereq {
		s, err := adb.ReadText(filepath.Join(plugins, prereq[i]))
		if err != nil {
			return err
		}
		if err = vm.runtime.DoStringNamed(s, prereq[i]); err != nil {
			return err
		}
	}
	return nil
}

func cleanModulePath(module string) (string, error) {
	module = strings.TrimSpace(module)
	if module == "" {
		return "", fmt.Errorf("module name is empty")
	}
	module = strings.TrimSuffix(module, ".lua")
	module = strings.ReplaceAll(module, "\\", "/")
	module = strings.TrimPrefix(module, "./")
	if module == "api" {
		return "", nil
	}
	if !strings.Contains(module, "/") {
		module = strings.ReplaceAll(module, ".", "/")
	}
	module += ".lua"
	if filepath.IsAbs(module) || filepath.VolumeName(module) != "" {
		return "", fmt.Errorf("absolute module paths are not allowed: %s", module)
	}
	clean := filepath.Clean(module)
	if clean == "." || strings.HasPrefix(clean, "..") ||
		strings.Contains(clean, string(filepath.Separator)+".."+string(filepath.Separator)) {
		return "", fmt.Errorf("module path escapes plugin root: %s", module)
	}
	return clean, nil
}

func (vm *LuaVM) setupRequire() error {
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		if state.Top() != 1 || !state.IsString(1) {
			return state.ArgError(1, "expected module name")
		}
		module := state.ToString(1)
		path, err := cleanModulePath(module)
		if err != nil {
			return state.Error(err.Error())
		}
		if path == "" {
			state.PushBoolean(true)
			return 1
		}
		data, err := vm.sandbox.ReadFile(path)
		if err != nil {
			return state.Error(fmt.Sprintf("failed to require %q: %v", module, err))
		}
		if err = state.LoadString(string(data), "@"+filepath.ToSlash(path)); err != nil {
			return state.Error(err.Error())
		}
		if err = state.Call(0, 1); err != nil {
			return state.Error(fmt.Sprintf("failed to execute module %q: %v", module, err))
		}
		return 1
	})
	vm.runtime.SetGlobal("__kaiju_load_module")
	return vm.runtime.DoStringNamed(`
local __kaiju_require_cache = {}
function require(name)
	if __kaiju_require_cache[name] ~= nil then
		return __kaiju_require_cache[name]
	end
	local loaded = __kaiju_load_module(name)
	if loaded == nil then loaded = true end
	__kaiju_require_cache[name] = loaded
	return loaded
end`, "kaiju require")
}

func launchPlugin(adb assets.Database, entry string) (*LuaVM, error) {
	defer tracing.NewRegion("plugins.launchPlugin").End()
	vm := &LuaVM{
		PluginPath: entry,
		runtime:    lua.New(),
	}
	if err := vm.runtime.OpenLibraries(); err != nil {
		return vm, err
	}
	if lua, err := os.ReadFile(entry); err == nil {
		root := filepath.Dir(entry)
		sandbox, err := os.OpenRoot(root)
		if err != nil {
			return vm, err
		}
		vm.sandbox = sandbox
		if err := vm.setupPrerequisites(adb); err != nil {
			return vm, err
		}
		if err := vm.setupRequire(); err != nil {
			return vm, err
		}
		vm.runtime.SandboxLibraries()
		for _, t := range reflectedTypes() {
			reflectStructToLua(t, vm)
		}
		if err := vm.runtime.DoStringNamed(string(lua), "@main.lua"); err != nil {
			return vm, err
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
		if err != nil {
			slog.Error("plugin failed to load", "plugin", dirs[i].Name(), "error", err)
			if vm != nil {
				vm.Close()
			}
			continue
		}
		vms = append(vms, vm)
	}
	return vms, nil
}
