/******************************************************************************/
/* editor_workspace_order_test.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"testing"

	"kaijuengine.com/editor/editor_settings"
)

func TestInsertDefaultWorkspaceConfigBeforeLaterDefaultWorkspace(t *testing.T) {
	workspaces := []editor_settings.WorkspaceConfig{
		{ID: "stage", Enabled: true},
		{ID: "settings", Enabled: true},
	}

	workspaces = insertDefaultWorkspaceConfig(workspaces, editor_settings.WorkspaceConfig{
		ID:      "schema",
		Enabled: true,
	})

	got := workspaceConfigIDs(workspaces)
	want := []string{"stage", "schema", "settings"}
	if !sliceEqual(got, want) {
		t.Fatalf("workspace order = %v, want %v", got, want)
	}
}

func TestInsertDefaultWorkspaceConfigAppendsUnknownWorkspace(t *testing.T) {
	workspaces := []editor_settings.WorkspaceConfig{
		{ID: "stage", Enabled: true},
		{ID: "settings", Enabled: true},
	}

	workspaces = insertDefaultWorkspaceConfig(workspaces, editor_settings.WorkspaceConfig{
		ID:      "plugin.test",
		Enabled: true,
	})

	got := workspaceConfigIDs(workspaces)
	want := []string{"stage", "settings", "plugin.test"}
	if !sliceEqual(got, want) {
		t.Fatalf("workspace order = %v, want %v", got, want)
	}
}

func workspaceConfigIDs(workspaces []editor_settings.WorkspaceConfig) []string {
	ids := make([]string, len(workspaces))
	for i := range workspaces {
		ids[i] = workspaces[i].ID
	}
	return ids
}
