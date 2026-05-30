/******************************************************************************/
/* history_stage_workspace_detach_entity_data.go                              */
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

type EntityDataDetachHistory struct {
	DetailsWorkspace *WorkspaceDetailsUI
	Entity           *editor_stage_manager.StageEntity
	Data             *entity_data_binding.EntityDataEntry
}

func (h *EntityDataDetachHistory) Redo() {
	defer tracing.NewRegion("EntityDataDetachHistory.Redo").End()
	h.Entity.DetachDataBinding(h.Data)
	w := h.DetailsWorkspace.workspace.Value()
	data_binding_renderer.Detatched(h.Data, weak.Make(w.Host), w.stageView.Manager(), h.Entity)
	w.hierarchyUI.refreshEntityBadgeForEntity(h.Entity)
	h.DetailsWorkspace.entitySelected(h.Entity)
}

func (h *EntityDataDetachHistory) Undo() {
	defer tracing.NewRegion("EntityDataDetachHistory.Undo").End()
	h.Entity.AttachDataBinding(h.Data)
	w := h.DetailsWorkspace.workspace.Value()
	data_binding_renderer.Attached(h.Data, weak.Make(w.Host), w.stageView.Manager(), h.Entity)
	w.hierarchyUI.refreshEntityBadgeForEntity(h.Entity)
	h.DetailsWorkspace.entitySelected(h.Entity)
}

func (h *EntityDataDetachHistory) Delete() {}
func (h *EntityDataDetachHistory) Exit()   {}
