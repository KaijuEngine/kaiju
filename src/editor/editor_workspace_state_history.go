/******************************************************************************/
/* editor_workspace_state_history.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import "kaijuengine.com/platform/profiler/tracing"

// workspaceStateHistory captures a workspace switch for undo/redo. setWorkspaceState
// already no-ops if the target id is no longer in activeWorkspaces, so an entry
// pointing at a workspace the user has since disabled silently does nothing
// rather than panicking.
type workspaceStateHistory struct {
	ed   *Editor
	from WorkspaceState
	to   WorkspaceState
}

func (h *workspaceStateHistory) Redo() {
	defer tracing.NewRegion("workspaceStateHistory.Redo").End()
	h.ed.setWorkspaceState(h.to)
}

func (h *workspaceStateHistory) Undo() {
	defer tracing.NewRegion("workspaceStateHistory.Undo").End()
	h.ed.setWorkspaceState(h.from)
}

func (h *workspaceStateHistory) Delete() {}
func (h *workspaceStateHistory) Exit()   {}
