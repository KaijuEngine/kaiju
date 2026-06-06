package plugins

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
)

type luaBridgeThing struct {
	Value int
}

func (t *luaBridgeThing) SetValue(v int) {
	t.Value = v
}

func (t *luaBridgeThing) Add(v matrix.Vec3) matrix.Vec3 {
	return v.Add(matrix.NewVec3(matrix.Float(t.Value), 0, 0))
}

func (t *luaBridgeThing) UsePointer(v *matrix.Vec3) matrix.Float {
	return v.X()
}

func (t *luaBridgeThing) Pointer() *matrix.Vec3 {
	v := matrix.NewVec3(4, 5, 6)
	return &v
}

func (t *luaBridgeThing) Slice() []int {
	return []int{1, 2, 3}
}

func (t *luaBridgeThing) Number() int {
	return 7
}

func testPluginDB() assets.Database {
	return assets.NewMockDB(map[string][]byte{
		filepath.Join(plugins, "debugger.lua"): []byte(`function breakpoint() end`),
		filepath.Join(plugins, "globals.lua"): []byte(`
function create_obj(self)
	local o = {}
	for k, v in pairs(self) do o[k] = v end
	setmetatable(o, { __index = self, __gc = self.__gc })
	return o
end`),
	})
}

func withTestRegistry(t *testing.T) {
	t.Helper()
	prev := GamePluginRegistry
	GamePluginRegistry = []reflect.Type{reflect.TypeFor[luaBridgeThing]()}
	t.Cleanup(func() { GamePluginRegistry = prev })
}

func writePlugin(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	pluginDir := filepath.Join(root, "test")
	if err := os.MkdirAll(pluginDir, os.ModePerm); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		path := filepath.Join(pluginDir, name)
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(pluginDir, "main.lua")
}

func TestLaunchPluginReflectionRoundTrip(t *testing.T) {
	withTestRegistry(t)
	entry := writePlugin(t, map[string]string{
		"helper.lua": `return { answer = 42 }`,
		"main.lua": `
local helper = require("helper")
local helperAgain = require("helper")
local thing = luaBridgeThing.New()
thing:SetValue(10)
local v = Vec3.New(1, 2, 3)
local added = thing:Add(v)
local p = thing:Pointer()
local s = thing:Slice()
result = helper.answer == 42
	and helper == helperAgain
	and added:X() == 11
	and added:Y() == 2
	and added:Z() == 3
	and thing:UsePointer(p) == 4
	and p:Z() == 6
	and s[1] == 1
	and s[2] == 2
	and s[3] == 3
	and thing:Number() == 7
`,
	})
	vm, err := launchPlugin(testPluginDB(), entry)
	if err != nil {
		t.Fatal(err)
	}
	if vm.runtime.PinnedPointerCount() == 0 {
		t.Fatal("expected reflected objects to be pinned for the VM lifetime")
	}
	defer vm.Close()

	vm.runtime.Global("result")
	defer vm.runtime.Pop(1)
	if !vm.runtime.IsBoolean(-1) || !vm.runtime.ToBoolean(-1) {
		t.Fatal("lua reflection round-trip result was false")
	}
}

func TestLaunchPluginSurfacesProtectedErrors(t *testing.T) {
	withTestRegistry(t)
	entry := writePlugin(t, map[string]string{
		"main.lua": `error("boom")`,
	})
	vm, err := launchPlugin(testPluginDB(), entry)
	if vm != nil {
		vm.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected protected lua error to surface, got %v", err)
	}
}

func TestLaunchPluginRejectsUnsafeRequire(t *testing.T) {
	withTestRegistry(t)
	entry := writePlugin(t, map[string]string{
		"main.lua": `require("../outside")`,
	})
	vm, err := launchPlugin(testPluginDB(), entry)
	if vm != nil {
		vm.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "escapes plugin root") {
		t.Fatalf("expected unsafe require error, got %v", err)
	}
}

func TestLaunchPluginBadArgumentReporting(t *testing.T) {
	withTestRegistry(t)
	entry := writePlugin(t, map[string]string{
		"main.lua": `
local thing = luaBridgeThing.New()
thing:SetValue("not a number")
`,
	})
	vm, err := launchPlugin(testPluginDB(), entry)
	if vm != nil {
		vm.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "SetValue argument 1") {
		t.Fatalf("expected reflected argument error, got %v", err)
	}
}
