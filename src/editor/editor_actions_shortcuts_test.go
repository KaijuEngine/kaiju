/******************************************************************************/
/* editor_actions_shortcuts_test.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"testing"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/editor/editor_workspace/terrain_workspace"
	"kaijuengine.com/platform/hid"
)

func TestBindableVariantActionsAreRegistered(t *testing.T) {
	ed := &Editor{}
	ed.history.Initialize(8)
	ed.initializeActions()
	checks := []struct {
		id       string
		visible  bool
		bindable bool
	}{
		{string(ActionStageSpawnPrimitive), false, false},
		{string(ActionStageSpawnCube), true, true},
		{string(ActionStageSpawnSphere), true, true},
		{string(ActionEditorBuild), false, false},
		{string(ActionEditorBuildDebug), true, true},
		{string(ActionEditorBuildRelease), true, true},
		{string(ActionEditorBuildAndRun), false, false},
		{string(ActionEditorRunCurrentStage), true, true},
	}
	for _, check := range checks {
		def, ok := ed.Actions().Registry().Definition(editor_action.ActionID(check.id))
		if !ok {
			t.Fatalf("action %s was not registered", check.id)
		}
		if def.Visible != check.visible {
			t.Fatalf("action %s visible = %v, want %v", check.id, def.Visible, check.visible)
		}
		if (def.Visible && !def.Unbindable) != check.bindable {
			t.Fatalf("action %s bindable = %v, want %v", check.id, def.Visible && !def.Unbindable, check.bindable)
		}
	}
}

func TestStageViewActionsDefaultBindings(t *testing.T) {
	ed := &Editor{}
	ed.history.Initialize(8)
	ed.initializeActions()
	checks := []struct {
		action editor_action.ActionID
		chord  editor_action.KeyChord
	}{
		{ActionStageToggleViewportLayout, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyP)}}},
		{ActionStageToggleContentPanel, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyC)}}},
		{ActionStageToggleHierarchyPanel, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyH)}}},
		{ActionStageToggleDetailsPanel, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyD)}}},
		{ActionStageRenameActor, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF2)}}},
		{ActionStageFocusSelection, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF)}}},
		{ActionStageTransformMove, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyW)}}},
		{ActionStageTransformRotate, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyE)}}},
		{ActionStageTransformScale, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyR)}}},
		{ActionStageWireframeMove, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyW)}, Alt: true}},
		{ActionStageWireframeRotate, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyE)}, Alt: true}},
		{ActionStageWireframeScale, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyR)}, Alt: true}},
	}
	for _, check := range checks {
		bindings := editor_action.BindingsForAction(
			ed.Actions().DefaultBindings(), nil, check.action, stage_workspace.ID)
		if len(bindings) != 1 {
			t.Fatalf("%s bindings = %d, want 1", check.action, len(bindings))
		}
		chord := bindings[0].Chord
		if !editor_action.ChordsEqual(chord, check.chord) {
			t.Fatalf("%s chord = %s, want %s", check.action,
				editor_action.FormatKeyChord(chord), editor_action.FormatKeyChord(check.chord))
		}
	}
}

func TestTerrainActionsDefaultBindings(t *testing.T) {
	ed := &Editor{}
	ed.history.Initialize(8)
	ed.initializeActions()
	checks := []struct {
		action editor_action.ActionID
		chord  editor_action.KeyChord
	}{
		{ActionTerrainDecreaseBrushRadius, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyOpenBracket)}}},
		{ActionTerrainIncreaseBrushRadius, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyCloseBracket)}}},
		{ActionTerrainDecreaseBrushStrength, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyOpenBracket)}, Shift: true}},
		{ActionTerrainIncreaseBrushStrength, editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyCloseBracket)}, Shift: true}},
	}
	for _, check := range checks {
		bindings := editor_action.BindingsForAction(
			ed.Actions().DefaultBindings(), nil, check.action, terrain_workspace.ID)
		if len(bindings) != 1 {
			t.Fatalf("%s bindings = %d, want 1", check.action, len(bindings))
		}
		chord := bindings[0].Chord
		if !editor_action.ChordsEqual(chord, check.chord) {
			t.Fatalf("%s chord = %s, want %s", check.action,
				editor_action.FormatKeyChord(chord), editor_action.FormatKeyChord(check.chord))
		}
	}
}
