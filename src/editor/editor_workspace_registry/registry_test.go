/******************************************************************************/
/* registry_test.go                                                           */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

/******************************************************************************/
/* registry_test.go                                                           */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/

package editor_workspace_registry_test

import (
	"testing"

	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
)

// fakeWorkspace is a minimal Workspace implementation used to exercise the
// registry without dragging in the editor or any built-in workspace's init().
type fakeWorkspace struct {
	id       string
	name     string
	required bool
}

func (f *fakeWorkspace) ID() string                                                 { return f.id }
func (f *fakeWorkspace) DisplayName() string                                        { return f.name }
func (f *fakeWorkspace) IsRequired() bool                                           { return f.required }
func (f *fakeWorkspace) Initialize(editor_workspace.WorkspaceEditorInterface) error { return nil }
func (f *fakeWorkspace) Shutdown()                                                  {}
func (f *fakeWorkspace) Open()                                                      {}
func (f *fakeWorkspace) Close()                                                     {}
func (f *fakeWorkspace) Focus()                                                     {}
func (f *fakeWorkspace) Blur()                                                      {}
func (f *fakeWorkspace) Hotkeys() []common_workspace.HotKey                         { return nil }
func (f *fakeWorkspace) Update(float64)                                             {}
func (f *fakeWorkspace) IsFocusedOnInput() bool                                     { return false }

func TestRegisterAndGet(t *testing.T) {
	w := &fakeWorkspace{id: "test_register_and_get", name: "RegGet"}
	editor_workspace_registry.Register(w)

	got, ok := editor_workspace_registry.Get("test_register_and_get")
	if !ok {
		t.Fatalf("expected workspace to be registered")
	}
	if got != w {
		t.Errorf("expected to get back the same workspace pointer")
	}
}

func TestGetMissing(t *testing.T) {
	if _, ok := editor_workspace_registry.Get("definitely_not_registered_id_xyz"); ok {
		t.Errorf("expected Get on unknown id to return false")
	}
}

func TestRegisterPreservesOrder(t *testing.T) {
	a := &fakeWorkspace{id: "test_order_a", name: "A"}
	b := &fakeWorkspace{id: "test_order_b", name: "B"}
	c := &fakeWorkspace{id: "test_order_c", name: "C"}
	editor_workspace_registry.Register(a)
	editor_workspace_registry.Register(b)
	editor_workspace_registry.Register(c)

	ids := editor_workspace_registry.IDs()
	pos := map[string]int{}
	for i, id := range ids {
		pos[id] = i
	}
	if _, ok := pos["test_order_a"]; !ok {
		t.Fatalf("test_order_a missing from IDs()")
	}
	if _, ok := pos["test_order_b"]; !ok {
		t.Fatalf("test_order_b missing from IDs()")
	}
	if _, ok := pos["test_order_c"]; !ok {
		t.Fatalf("test_order_c missing from IDs()")
	}
	if pos["test_order_a"] >= pos["test_order_b"] {
		t.Errorf("expected a before b: a=%d b=%d", pos["test_order_a"], pos["test_order_b"])
	}
	if pos["test_order_b"] >= pos["test_order_c"] {
		t.Errorf("expected b before c: b=%d c=%d", pos["test_order_b"], pos["test_order_c"])
	}
}

func TestDuplicateRegisterIgnored(t *testing.T) {
	first := &fakeWorkspace{id: "test_duplicate", name: "First"}
	second := &fakeWorkspace{id: "test_duplicate", name: "Second"}
	editor_workspace_registry.Register(first)
	editor_workspace_registry.Register(second) // logs error, does not replace

	got, ok := editor_workspace_registry.Get("test_duplicate")
	if !ok {
		t.Fatalf("expected first registration to remain")
	}
	if got != first {
		t.Errorf("duplicate id must not overwrite the original registration")
	}
}

func TestEmptyIDIgnored(t *testing.T) {
	w := &fakeWorkspace{id: "", name: "empty"}
	editor_workspace_registry.Register(w)
	if _, ok := editor_workspace_registry.Get(""); ok {
		t.Errorf("empty-id workspace must not be registered")
	}
}
