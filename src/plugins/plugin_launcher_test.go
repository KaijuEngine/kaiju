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

type luaBridgeTestObject struct {
	Value float64
	Vec   matrix.Vec3
	Child *luaBridgeTestObject
}

func (o *luaBridgeTestObject) NewChild(value float64) *luaBridgeTestObject {
	o.Child = &luaBridgeTestObject{Value: value}
	return o.Child
}

func (o *luaBridgeTestObject) CopyValueFrom(other *luaBridgeTestObject) {
	o.Value = other.Value
}

func (o *luaBridgeTestObject) SetVec(v matrix.Vec3) {
	o.Vec = v
}

func (o *luaBridgeTestObject) FailIfCalledWithMissingArg(v float64) float64 {
	return v
}

func testAssetDB(t *testing.T) assets.Database {
	t.Helper()
	adb, err := assets.NewFileDatabase("../editor/editor_embedded_content/editor_content")
	if err != nil {
		t.Fatal(err)
	}
	return adb
}

func writeLuaTestScript(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "main.lua")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLuaBridgePointerReturnCanBePassedBack(t *testing.T) {
	root := &luaBridgeTestObject{}
	script := writeLuaTestScript(t, `
function main(root)
	local child = root:NewChild(42)
	root:CopyValueFrom(child)
end
`)
	vm, err := LaunchScript(testAssetDB(t), script,
		[]reflect.Type{reflect.TypeFor[luaBridgeTestObject]()},
		map[string]reflect.Value{"root": reflect.ValueOf(root)})
	if vm != nil {
		defer vm.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.InvokeGlobalFunctionWithArgs("main", reflect.ValueOf(root)); err != nil {
		t.Fatal(err)
	}
	if root.Value != 42 {
		t.Fatalf("root.Value = %v, want 42", root.Value)
	}
}

func TestLuaBridgeVec3ConstructorPassesValue(t *testing.T) {
	root := &luaBridgeTestObject{}
	script := writeLuaTestScript(t, `
function main(root)
	root:SetVec(Vec3.New(1, 2, 3))
end
`)
	vm, err := LaunchScript(testAssetDB(t), script,
		[]reflect.Type{reflect.TypeFor[luaBridgeTestObject]()},
		map[string]reflect.Value{"root": reflect.ValueOf(root)})
	if vm != nil {
		defer vm.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.InvokeGlobalFunctionWithArgs("main", reflect.ValueOf(root)); err != nil {
		t.Fatal(err)
	}
	if root.Vec != (matrix.Vec3{1, 2, 3}) {
		t.Fatalf("root.Vec = %v, want [1 2 3]", root.Vec)
	}
}

func TestLuaBridgeReportsWrongArgumentCount(t *testing.T) {
	root := &luaBridgeTestObject{}
	script := writeLuaTestScript(t, `
function main(root)
	root:FailIfCalledWithMissingArg()
end
`)
	vm, err := LaunchScript(testAssetDB(t), script,
		[]reflect.Type{reflect.TypeFor[luaBridgeTestObject]()},
		map[string]reflect.Value{"root": reflect.ValueOf(root)})
	if vm != nil {
		defer vm.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	err = vm.InvokeGlobalFunctionWithArgs("main", reflect.ValueOf(root))
	if err == nil || !strings.Contains(err.Error(), "expects 1 arguments") {
		t.Fatalf("expected wrong argument error, got %v", err)
	}
}

func TestLuaLoaderLoadsRequiredFilesAndSandboxesOS(t *testing.T) {
	root := &luaBridgeTestObject{}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "helper.lua"), []byte(`
function helper_value()
	return 7
end
`), 0644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "main.lua")
	if err := os.WriteFile(script, []byte(`
require("helper")
function main(root)
	if os ~= nil then error("os should be sandboxed") end
	local child = root:NewChild(helper_value())
	root:CopyValueFrom(child)
end
`), 0644); err != nil {
		t.Fatal(err)
	}
	vm, err := LaunchScript(testAssetDB(t), script,
		[]reflect.Type{reflect.TypeFor[luaBridgeTestObject]()},
		map[string]reflect.Value{"root": reflect.ValueOf(root)})
	if vm != nil {
		defer vm.Close()
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.InvokeGlobalFunctionWithArgs("main", reflect.ValueOf(root)); err != nil {
		t.Fatal(err)
	}
	if root.Value != 7 {
		t.Fatalf("root.Value = %v, want 7", root.Value)
	}
}
