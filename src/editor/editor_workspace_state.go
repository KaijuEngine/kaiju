package editor

type WorkspaceState = uint8

const (
	WorkspaceStateNone = WorkspaceState(iota)
	WorkspaceStateStage
	WorkspaceStateContent
)

func (ed *Editor) setWorkspaceState(state WorkspaceState) {
	if ed.workspaceState == state {
		return
	}
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Close()
	}
	ed.workspaceState = state
	switch ed.workspaceState {
	case WorkspaceStateStage:
		ed.currentWorkspace = &ed.workspaces.stage
	case WorkspaceStateContent:
		ed.currentWorkspace = &ed.workspaces.content
	}
	ed.currentWorkspace.Open()
}
