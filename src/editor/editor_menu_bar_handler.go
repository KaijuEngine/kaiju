package editor

// OnStageWorkspaceSelected will inform the editor that the developer has
// changed to the stage workspace. This is an exposed function to meet the
// interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) OnStageWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateStage)
}

// OnContentWorkspaceSelected will inform the editor that the developer has
// changed to the content workspace. This is an exposed function to meet the
// interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) OnContentWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateContent)
}
