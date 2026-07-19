package stage_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine/stages"
)

type entityGIContributionHistory struct {
	ui     *WorkspaceDetailsUI
	entity *editor_stage_manager.StageEntity
	from   stages.GIContribution
	to     stages.GIContribution
}

func (h *entityGIContributionHistory) apply(value stages.GIContribution) {
	if h.entity == nil {
		return
	}
	h.entity.StageData.Description.GIContribution = value
	if h.ui != nil {
		h.ui.reload()
	}
}

func (h *entityGIContributionHistory) Redo() { h.apply(h.to) }
func (h *entityGIContributionHistory) Undo() { h.apply(h.from) }
func (*entityGIContributionHistory) Delete() {}
func (*entityGIContributionHistory) Exit()   {}
