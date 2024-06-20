/******************************************************************************/
/* codegen_test.go                                                            */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
