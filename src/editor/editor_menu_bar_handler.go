package editor

func (ed *Editor) OnStageWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateStage)
}

func (ed *Editor) OnContentWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateContent)
}
