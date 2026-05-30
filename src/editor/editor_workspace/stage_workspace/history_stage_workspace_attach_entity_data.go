/******************************************************************************/
/* history_stage_workspace_attach_entity_data.go                              */
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

type EntityDataAttachHistory struct {
	DetailsWorkspace *WorkspaceDetailsUI
	Entity           *editor_stage_manager.StageEntity
	Data             *entity_data_binding.EntityDataEntry
}

func (h *EntityDataAttachHistory) Redo() {
	defer tracing.NewRegion("EntityDataAttachHistory.Redo").End()
	h.Entity.AttachDataBinding(h.Data)
	w := h.DetailsWorkspace.workspace.Value()
	data_binding_renderer.Attached(h.Data, weak.Make(w.Host), w.stageView.Manager(), h.Entity)
	w.hierarchyUI.refreshEntityBadgeForEntity(h.Entity)
	h.DetailsWorkspace.entitySelected(h.Entity)
}

func (h *EntityDataAttachHistory) Undo() {
	defer tracing.NewRegion("EntityDataAttachHistory.Undo").End()
	h.Entity.DetachDataBinding(h.Data)
	w := h.DetailsWorkspace.workspace.Value()
	data_binding_renderer.Detatched(h.Data, weak.Make(w.Host), w.stageView.Manager(), h.Entity)
	w.hierarchyUI.refreshEntityBadgeForEntity(h.Entity)
	h.DetailsWorkspace.entitySelected(h.Entity)
}

func (h *EntityDataAttachHistory) Delete() {}
func (h *EntityDataAttachHistory) Exit()   {}
