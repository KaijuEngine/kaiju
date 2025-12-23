package stage_workspace

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/data_binding_renderer"
	"kaiju/platform/profiler/tracing"
	"weak"
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
	h.DetailsWorkspace.entitySelected(h.Entity)
}

func (h *EntityDataDetachHistory) Undo() {
	defer tracing.NewRegion("EntityDataDetachHistory.Undo").End()
	h.Entity.AttachDataBinding(h.Data)
	w := h.DetailsWorkspace.workspace.Value()
	data_binding_renderer.Attached(h.Data, weak.Make(w.Host), w.stageView.Manager(), h.Entity)
	h.DetailsWorkspace.entitySelected(h.Entity)
}

func (h *EntityDataDetachHistory) Delete() {}
func (h *EntityDataDetachHistory) Exit()   {}
