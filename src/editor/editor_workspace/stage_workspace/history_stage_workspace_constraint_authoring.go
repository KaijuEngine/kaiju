/******************************************************************************/
/* history_stage_workspace_constraint_authoring.go                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"weak"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/platform/profiler/tracing"
)

type constraintDataAttachHistory struct {
	workspace *StageWorkspace
	Entity    *editor_stage_manager.StageEntity
	Data      *entity_data_binding.EntityDataEntry
}

func (h *constraintDataAttachHistory) Redo() {
	defer tracing.NewRegion("constraintDataAttachHistory.Redo").End()
	man := h.workspace.stageView.Manager()
	h.Entity.AttachDataBinding(h.Data)
	data_binding_renderer.Attached(h.Data, weak.Make(h.workspace.Host), man, h.Entity)
	if man.IsSelected(h.Entity) {
		data_binding_renderer.ShowSpecific(h.Data, weak.Make(h.workspace.Host), h.Entity)
	}
	h.workspace.hierarchyUI.refreshEntityBadgeForEntity(h.Entity)
	h.workspace.detailsUI.reload()
}

func (h *constraintDataAttachHistory) Undo() {
	defer tracing.NewRegion("constraintDataAttachHistory.Undo").End()
	man := h.workspace.stageView.Manager()
	h.Entity.DetachDataBinding(h.Data)
	data_binding_renderer.Detatched(h.Data, weak.Make(h.workspace.Host), man, h.Entity)
	h.workspace.hierarchyUI.refreshEntityBadgeForEntity(h.Entity)
	h.workspace.detailsUI.reload()
}

func (h *constraintDataAttachHistory) Delete() {}
func (h *constraintDataAttachHistory) Exit()   {}
