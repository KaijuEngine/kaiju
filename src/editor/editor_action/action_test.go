/******************************************************************************/
/* action_test.go                                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_action

import (
	"encoding/json"
	"errors"
	"testing"

	"kaijuengine.com/platform/hid"
)

type testPrimitiveParams struct {
	Primitive string `json:"primitive"`
}

type testValueParams struct {
	Value string `json:"value"`
}

func TestRegistryRejectsDuplicateActions(t *testing.T) {
	registry := NewRegistry()
	def := Definition{ID: "test.action", Label: "Test Action", Visible: true}
	if err := registry.Register(def, func(Context, Request) Result {
		return Success("")
	}, nil); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	err := registry.Register(def, func(Context, Request) Result {
		return Success("")
	}, nil)
	if !errors.Is(err, ErrDuplicateAction) {
		t.Fatalf("duplicate error = %v, want %v", err, ErrDuplicateAction)
	}
}

func TestServiceSearchMatchesVariantsAndFiltersCanRun(t *testing.T) {
	service := NewService()
	if err := service.Register(Definition{
		ID:            "stage.spawnPrimitive",
		Label:         "Spawn Primitive",
		Category:      "Stage",
		Tags:          []string{"mesh"},
		Visible:       true,
		DefaultParams: Params(testPrimitiveParams{Primitive: "cube"}),
		NewParams:     func() any { return &testPrimitiveParams{} },
		Variants: []Variant{
			{Label: "Spawn Cube", Tags: []string{"cube"}, Params: Params(testPrimitiveParams{Primitive: "cube"})},
			{Label: "Spawn Sphere", Tags: []string{"sphere"}, Params: Params(testPrimitiveParams{Primitive: "sphere"})},
		},
	}, func(Context, Request) Result {
		return Success("")
	}, func(_ Context, req Request) Result {
		params, _ := Param[testPrimitiveParams](req)
		if params.Primitive == "sphere" {
			return Failure("sphere unavailable")
		}
		return Success("")
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	entries := service.Search("cube")
	if len(entries) != 1 {
		t.Fatalf("cube search returned %d entries, want 1", len(entries))
	}
	if entries[0].Label != "Spawn Cube" {
		t.Fatalf("cube search label = %q, want Spawn Cube", entries[0].Label)
	}
	entries = service.Search("sphere")
	if len(entries) != 0 {
		t.Fatalf("sphere search returned %d entries, want 0 because CanRun rejects it", len(entries))
	}
}

func TestServiceRunWrapsTransactionalActions(t *testing.T) {
	service := NewService()
	var begin, commit int
	service.SetTransactionHooks(func() { begin++ }, func() { commit++ }, nil)
	if err := service.Register(Definition{
		ID:         "test.transaction",
		Label:      "Transaction",
		UndoPolicy: UndoPolicyTransaction,
		Visible:    true,
	}, func(Context, Request) Result {
		return Success("done")
	}, nil); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	result := service.Run(Request{ID: "test.transaction"})
	if !result.OK {
		t.Fatalf("run failed: %v", result)
	}
	if begin != 1 || commit != 1 {
		t.Fatalf("transaction hooks begin=%d commit=%d, want 1/1", begin, commit)
	}
}

func TestServiceRunOnMainThreadUsesScheduler(t *testing.T) {
	service := NewService()
	service.SetMainThreadScheduler(func(call func()) {
		call()
	})
	if err := service.Register(Definition{
		ID:        "test.params",
		Label:     "Params",
		Visible:   true,
		NewParams: func() any { return &testValueParams{} },
	}, func(_ Context, req Request) Result {
		params, ok := Param[testValueParams](req)
		if !ok {
			return Failure("missing params")
		}
		return Result{OK: true, Data: map[string]any{"value": params.Value}}
	}, nil); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	result := service.RunOnMainThread(Request{
		ID:     "test.params",
		Params: Params(testValueParams{Value: "ok"}),
	})
	if !result.OK || result.Data["value"] != "ok" {
		t.Fatalf("result = %#v, want ok data", result)
	}
}

func TestServiceRunParsesJSONParamsToConcreteStruct(t *testing.T) {
	service := NewService()
	if err := service.Register(Definition{
		ID:        "test.jsonParams",
		Label:     "JSON Params",
		Visible:   true,
		NewParams: func() any { return &testValueParams{} },
	}, func(_ Context, req Request) Result {
		params, ok := Param[testValueParams](req)
		if !ok {
			return Failure("missing params")
		}
		return Result{OK: true, Data: map[string]any{"value": params.Value}}
	}, nil); err != nil {
		t.Fatalf("register failed: %v", err)
	}
	result := service.Run(Request{
		ID:     "test.jsonParams",
		Params: json.RawMessage(`{"value":"from-json"}`),
	})
	if !result.OK || result.Data["value"] != "from-json" {
		t.Fatalf("result = %#v, want parsed JSON params", result)
	}
}

func TestEffectiveBindingsPreservesDefaultsAndAllowsOverrides(t *testing.T) {
	action := ActionID("editor.redo")
	defaults := []ActionBinding{
		{Action: action, Enabled: true, Chord: KeyChord{Keys: []int{int(hid.KeyboardKeyY)}, CtrlOrMeta: true}},
		{Action: action, Enabled: true, Chord: KeyChord{Keys: []int{int(hid.KeyboardKeyZ)}, CtrlOrMeta: true, Shift: true}},
	}
	effective := EffectiveBindings(defaults, nil)
	if len(effective) != 2 {
		t.Fatalf("effective default bindings = %d, want 2", len(effective))
	}
	override := []ActionBinding{
		{Action: action, Enabled: true, Chord: KeyChord{Keys: []int{int(hid.KeyboardKeyR)}, CtrlOrMeta: true}},
	}
	effective = EffectiveBindings(defaults, override)
	if len(effective) != 1 {
		t.Fatalf("effective overridden bindings = %d, want 1", len(effective))
	}
	if !ChordsEqual(effective[0].Chord, override[0].Chord) {
		t.Fatalf("effective chord = %#v, want override %#v", effective[0].Chord, override[0].Chord)
	}
}

func TestEffectiveBindingsSupportsTombstones(t *testing.T) {
	action := ActionID("editor.saveStage")
	defaults := []ActionBinding{
		{Action: action, Enabled: true, Chord: KeyChord{Keys: []int{int(hid.KeyboardKeyS)}, CtrlOrMeta: true}},
	}
	user := []ActionBinding{{Action: action, Enabled: false}}
	effective := EffectiveBindings(defaults, user)
	if len(effective) != 0 {
		t.Fatalf("effective bindings = %d, want tombstone to clear defaults", len(effective))
	}
}

func TestBindingConflictsRespectScopes(t *testing.T) {
	chord := KeyChord{Keys: []int{int(hid.KeyboardKeyD)}, CtrlOrMeta: true}
	global := ActionBinding{Action: "editor.saveStage", Enabled: true, Chord: chord}
	stage := ActionBinding{Action: "stage.duplicateSelection", Workspace: "stage", Enabled: true, Chord: chord}
	content := ActionBinding{Action: "content.test", Workspace: "content", Enabled: true, Chord: chord}
	if conflicts := BindingConflicts([]ActionBinding{global}, stage, nil); len(conflicts) != 1 {
		t.Fatalf("workspace vs global conflicts = %d, want 1", len(conflicts))
	}
	if conflicts := BindingConflicts([]ActionBinding{stage}, content, nil); len(conflicts) != 0 {
		t.Fatalf("different workspace conflicts = %d, want 0", len(conflicts))
	}
	if conflicts := BindingConflicts([]ActionBinding{stage, content}, global, nil); len(conflicts) != 2 {
		t.Fatalf("global vs workspaces conflicts = %d, want 2", len(conflicts))
	}
}

func TestRemoveBindingConflictsCreatesDefaultTombstone(t *testing.T) {
	oldAction := ActionID("editor.saveStage")
	chord := KeyChord{Keys: []int{int(hid.KeyboardKeyS)}, CtrlOrMeta: true}
	defaults := []ActionBinding{{Action: oldAction, Enabled: true, Chord: chord}}
	conflicts := []BindingConflict{{Binding: defaults[0], Label: "Save Stage"}}
	user := RemoveBindingConflicts(defaults, nil, conflicts)
	if len(user) != 1 {
		t.Fatalf("user bindings = %d, want tombstone", len(user))
	}
	if user[0].Action != oldAction || user[0].Enabled {
		t.Fatalf("tombstone = %#v, want disabled old action", user[0])
	}
	if effective := EffectiveBindings(defaults, user); len(effective) != 0 {
		t.Fatalf("effective bindings = %d, want default removed", len(effective))
	}
}
