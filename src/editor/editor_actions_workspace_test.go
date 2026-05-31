/******************************************************************************/
/* editor_actions_workspace_test.go                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"testing"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/editor_workspace/common_workspace"
	"kaijuengine.com/editor/editor_workspace_registry"
)

type workspaceActionTestWorkspace struct {
	id     string
	name   string
	opened bool
}

func (w *workspaceActionTestWorkspace) ID() string          { return w.id }
func (w *workspaceActionTestWorkspace) DisplayName() string { return w.name }
func (w *workspaceActionTestWorkspace) IsRequired() bool    { return false }
func (w *workspaceActionTestWorkspace) Initialize(editor_workspace.WorkspaceEditorInterface) error {
	return nil
}
func (w *workspaceActionTestWorkspace) Shutdown()                          {}
func (w *workspaceActionTestWorkspace) Open()                              { w.opened = true }
func (w *workspaceActionTestWorkspace) Close()                             {}
func (w *workspaceActionTestWorkspace) Focus()                             {}
func (w *workspaceActionTestWorkspace) Blur()                              {}
func (w *workspaceActionTestWorkspace) Hotkeys() []common_workspace.HotKey { return nil }
func (w *workspaceActionTestWorkspace) Update(float64)                     {}
func (w *workspaceActionTestWorkspace) IsFocusedOnInput() bool             { return false }

func TestWorkspaceActionRegisteredForRegisteredWorkspace(t *testing.T) {
	workspace := &workspaceActionTestWorkspace{
		id:   "test_atomic_workspace_action",
		name: "Atomic Workspace",
	}
	editor_workspace_registry.Register(workspace)
	ed := &Editor{
		activeWorkspaces: map[string]editor_workspace.Workspace{
			workspace.id: workspace,
		},
	}
	ed.history.Initialize(8)
	ed.initializeActions()

	actionID := openWorkspaceActionID(workspace.id)
	def, ok := ed.Actions().Registry().Definition(actionID)
	if !ok {
		t.Fatalf("workspace action %q was not registered", actionID)
	}
	if !def.Visible {
		t.Fatalf("workspace action %q should be visible", actionID)
	}
	entries := ed.Actions().Search("open atomic workspace")
	found := false
	for _, entry := range entries {
		if entry.ID == actionID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("workspace action %q was not returned by search", actionID)
	}
	result := ed.Actions().Run(editor_action.Request{ID: actionID})
	if !result.OK {
		t.Fatalf("workspace action failed: %#v", result)
	}
	if ed.workspaceState != workspace.id {
		t.Fatalf("workspace state = %q, want %q", ed.workspaceState, workspace.id)
	}
	if !workspace.opened {
		t.Fatalf("workspace Open was not called")
	}
}
