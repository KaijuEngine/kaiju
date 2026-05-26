/******************************************************************************/
/* editor_workspace_state.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import "kaijuengine.com/platform/profiler/tracing"

// WorkspaceState is the id of the currently active workspace. Empty string
// means "none active". The set of valid values is determined at runtime by
// what's in the workspace registry, not by a hard-coded enum.
type WorkspaceState = string

const WorkspaceStateNone WorkspaceState = ""

// setWorkspaceState switches the editor to the workspace identified by state.
// No-ops if state matches the current state, or if state is unknown to the
// active workspace set (e.g. workspace was disabled, or never registered).
// Adds an undo entry so the user can navigate back via Ctrl+Z.
func (ed *Editor) setWorkspaceState(state WorkspaceState) {
	defer tracing.NewRegion("Editor.setWorkspaceState").End()
	if ed.workspaceState == state {
		return
	}
	next, ok := ed.activeWorkspaces[state]
	if !ok {
		return
	}
	if ed.workspaceState != WorkspaceStateNone {
		ed.history.Add(&workspaceStateHistory{
			ed:   ed,
			from: ed.workspaceState,
			to:   state,
		})
	}
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Close()
	}
	ed.workspaceState = state
	ed.currentWorkspace = next
	ed.globalInterfaces.menuBar.SetActiveTab(state)
	ed.currentWorkspace.Open()
}
