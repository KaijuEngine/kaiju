/******************************************************************************/
/* plugin_launcher.go                                                         */
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

	"kaiju/engine/assets"
	"kaiju/platform/profiler/tracing"
	"kaiju/plugins/lua"
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

func (vm *LuaVM) InvokeGlobalFunction(name string) {
	// TODO:  Support arguments...
	vm.runtime.Global(name)
	if vm.runtime.IsFunction(-1) {
		vm.runtime.Call(0, 0)
	} else {
		vm.runtime.Pop(1)
	}
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
	prereq := []string{"debugger.lua", "globals.lua"}
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

func launchPlugin(adb assets.Database, entry string) (*LuaVM, error) {
	defer tracing.NewRegion("plugins.launchPlugin").End()
	vm := &LuaVM{
		PluginPath: entry,
		runtime:    lua.New(),
	}
	vm.runtime.OpenLibraries()
	if err := vm.setupPrerequisites(adb); err != nil {
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
