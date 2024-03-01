package codegen

import (
	"bytes"
	"encoding/gob"
	"kaiju/engine"
	"reflect"
	"testing"
)

func TestWalk(t *testing.T) {
	gens, err := walkInternal("test_data", "test_data", ".txt")
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
	if gens[0].PkgPath != "test_data/sub_test_data" {
		t.Error("Expected first type to be in test_data, got ", gens[0].Pkg)
	}
	if gens[1].PkgPath != "test_data" {
		t.Error("Expected second type to be in test_data, got ", gens[1].Pkg)
	}
	rt := gens[0].New()
	rt.Value.Elem().FieldByName("Age").SetInt(10)
	thing := []engine.EntityData{rt.Value}
	s := bytes.NewBuffer([]byte{})
	enc := gob.NewEncoder(s)
	if err := enc.Encode(thing); err != nil {
		t.Error(err)
	}
	rt.Value.Elem().FieldByName("Age").SetInt(15)
	dec := gob.NewDecoder(s)
	out := []engine.EntityData{}
	if err := dec.Decode(&out); err != nil {
		t.Error(err)
	}
	v := reflect.ValueOf(out[0])
	a := v.Elem().FieldByName("Age").Int()
	if a != 10 {
		t.Error("Expected 10, got ", a)
	}
}
