package stage_workspace

import "kaijuengine.com/engine/stages"

type stageGISettingsHistory struct {
	ui   *WorkspaceGIUI
	from stages.StageGlobalIllumination
	to   stages.StageGlobalIllumination
}

func (h *stageGISettingsHistory) Redo() { h.ui.apply(h.to, false) }
func (h *stageGISettingsHistory) Undo() { h.ui.apply(h.from, false) }
func (*stageGISettingsHistory) Delete() {}
func (*stageGISettingsHistory) Exit()   {}
