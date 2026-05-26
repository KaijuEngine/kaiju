/******************************************************************************/
/* codegen_test.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package codegen

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"kaijuengine.com/engine/encoding/pod"
)

func TestWalk(t *testing.T) {
	srcRoot, err := os.OpenRoot("test_structure")
	if err != nil {
		t.Error(err)
	}
	gens, err := walkInternal(srcRoot, srcRoot, "test_structure", ".txt")
	if err != nil {
		t.Error(err)
	}
	if len(gens) != 2 {
		t.Error("Expected 2 generated types, got ", len(gens))
	}
	if gens[0].Name != "Nothing" {
		t.Error("Expected first type to be Nothing, got ", gens[0].Name)
	}
	if gens[1].Name != "SomeThing" {
		t.Error("Expected second type to be Something, got ", gens[1].Name)
	}
	if gens[0].Pkg != "sub_test_data" {
		t.Error("Expected first type to be in test_data, got ", gens[0].Pkg)
	}
	if gens[1].Pkg != "test_data" {
		t.Error("Expected second type to be in test_data, got ", gens[1].Pkg)
	}
	if gens[0].PkgPath != "test_structure/test_data/sub_test_data" {
		t.Error("Expected first type to be in test_data, got ", gens[0].Pkg)
	}
	if gens[1].PkgPath != "test_structure/test_data" {
		t.Error("Expected second type to be in test_data, got ", gens[1].Pkg)
	}
	rt := gens[0].New()
	rt.Value.Elem().FieldByName("Age").SetInt(10)
	thing := []any{rt.Value.Interface()}
	s := bytes.NewBuffer([]byte{})
	enc := pod.NewEncoder(s)
	if err := enc.Encode(thing); err != nil {
		t.Error(err)
	}
	dec := pod.NewDecoder(s)
	out := []any{}
	if err := dec.Decode(&out); err != nil {
		t.Error(err)
	}
	r := reflect.ValueOf(out[0])
	if r.Kind() == reflect.Pointer {
		r = r.Elem()
	}
	a := r.FieldByName("Age").Int()
	if a != 10 {
		t.Error("Expected 10, got ", a)
	}
}

func TestWalkRigidBodyShapeEnumIncludesTerrain(t *testing.T) {
	srcPath, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	readPath, err := filepath.Abs("../../engine_entity_data")
	if err != nil {
		t.Fatal(err)
	}
	srcRoot, err := os.OpenRoot(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	defer srcRoot.Close()
	readRoot, err := os.OpenRoot(readPath)
	if err != nil {
		t.Fatal(err)
	}
	defer readRoot.Close()
	gens, err := Walk(srcRoot, readRoot, "kaijuengine.com")
	if err != nil {
		t.Fatal(err)
	}
	var rigidBody GeneratedType
	for i := range gens {
		if gens[i].Name == "RigidBodyEntityData" {
			rigidBody = gens[i]
			break
		}
	}
	if !rigidBody.IsValid() {
		t.Fatal("expected generated RigidBodyEntityData")
	}
	for i := range rigidBody.Fields {
		if rigidBody.Fields[i].Name != "Shape" {
			continue
		}
		if len(rigidBody.FieldGens[i].EnumValues) == 0 {
			t.Fatal("expected Shape field to have enum metadata")
		}
		if got := rigidBody.FieldGens[i].EnumValues["ShapeTerrain"]; got != int64(6) {
			t.Fatalf("expected ShapeTerrain enum value 6, got %v", got)
		}
		return
	}
	t.Fatal("expected RigidBodyEntityData.Shape field")
}
