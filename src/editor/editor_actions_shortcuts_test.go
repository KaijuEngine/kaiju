/******************************************************************************/
/* editor_actions_shortcuts_test.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"testing"

	"kaijuengine.com/editor/editor_action"
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
