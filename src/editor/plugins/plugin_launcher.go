//go:build editor

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

	"kaiju/editor/editor_interface"
	"kaiju/editor/plugins/lua"
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
	requireReg = regexp.MustCompile(`require\s{0,}\({0,1}\s{0,}["'](.*?)(\.lau){0,}["']\s{0,}\){0,1}`)
)

type LuaVM struct {
	PluginPath string
	runtime    lua.State
	sandbox    *os.Root
}

func reflectStructToLua(t reflect.Type, vm *LuaVM) {
	name := t.Name()
	vm.runtime.NewTable()
	vm.runtime.PushGoFunction(func(state *lua.State) int {
		var to reflect.Value
		if state.IsUserData(-1) {
			to = reflect.ValueOf(state.ToUserData(-1))
		} else {
			to = reflect.New(t)
		}
		state.Global("create_obj")
		state.Global(name)
		state.PushBoolean(false)
		state.Call(2, 1)
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
			// TODO:  Validate the inputs
			state.Field(1, goPtrField)
			obj := state.ToUserData(-1)
			v := reflect.ValueOf(obj)
			args := make([]reflect.Value, argCount-1)
			for i := range argCount - 1 {
				idx := 2 + i
				if state.IsBoolean(idx) {
					args[i] = reflect.ValueOf(state.ToBoolean(idx))
				} else if state.IsNumber(idx) {
					n := state.ToNumber(idx)
					args[i] = reflect.ValueOf(n)
					args[i] = args[i].Convert(argTypes[i])
				} else if state.IsTable(idx) {
					state.Field(idx, goPtrField)
					args[i] = reflect.ValueOf(state.ToUserData(1))
				} else if state.IsUserData(idx) {
					args[i] = reflect.ValueOf(state.ToUserData(idx))
				} else if state.IsString(idx) {
					str := state.ToString(idx)
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
					rp := reflect.New(r.Type())
					rp.Elem().Set(r)
					state.Global(rt.Name())
					state.Field(-1, "New")
					state.PushUserData(rp)
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
	vm.runtime.Global(globalCleanupPtrFn)
	vm.runtime.SetField(-2, "__gc")
	vm.runtime.SetGlobal(name)
}

func removeImportAPI(lua string) string {
	return apiReg.ReplaceAllString(lua, "")
}

func (vm *LuaVM) rollup(lua, luaPath string, imported *[]string) error {
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

func (vm *LuaVM) setupPrerequisites(ed editor_interface.Editor) error {
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
	prereq := []string{"debugger.lua", "globals.lua"}
	for i := range prereq {
		err := vm.runtime.DoFile(ed.Host().AssetDatabase().ToRawPath(
			filepath.Join(plugins, prereq[i])))
		if err != nil {
			return err
		}
	}
	return nil
}

func launchPlugin(ed editor_interface.Editor, entry string) (*LuaVM, error) {
	vm := &LuaVM{
		PluginPath: entry,
		runtime:    lua.New(),
	}
	vm.runtime.OpenLibraries()
	if err := vm.setupPrerequisites(ed); err != nil {
		return vm, err
	}
	for _, t := range reflectedTypes() {
		reflectStructToLua(t, vm)
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
		// TODO:  Don't ignore this error
		wd, _ := os.Getwd()
		os.Chdir(vm.sandbox.Name())
		for i := range imports {
			refined := strings.TrimPrefix(filepath.ToSlash(
				strings.TrimPrefix(imports[i], vm.sandbox.Name())), "/")
			// TODO:  Can we just load up the file and not do it yet?
			vm.runtime.DoFile(filepath.ToSlash(refined))
		}
		// TODO:  Don't ignore this error
		os.Chdir(wd)
	} else {
		return vm, err
	}
	return vm, nil
}

func LaunchPlugins(ed editor_interface.Editor) ([]*LuaVM, error) {
	pluginsPath := filepath.Join("content", plugins)
	dirs, err := os.ReadDir(pluginsPath)
	vms := make([]*LuaVM, 0)
	if err != nil {
		return vms, err
	}
	for i := range dirs {
		if !dirs[i].IsDir() {
			continue
		}
		vm, err := launchPlugin(ed, filepath.Join(pluginsPath, dirs[i].Name(), "main.lua"))
		vms = append(vms, vm)
		if err != nil {
			slog.Error("plugin failed to load", "plugin", dirs[i].Name(), "error", err)
		}
	}
	return vms, nil
}
