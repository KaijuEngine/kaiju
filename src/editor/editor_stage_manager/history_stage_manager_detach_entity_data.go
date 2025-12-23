package editor_stage_manager

import (
	"kaiju/editor/codegen/entity_data_binding"
	"kaiju/platform/profiler/tracing"
)

type EntityDataDetachHistory struct {
	Entity *StageEntity
	Data   *entity_data_binding.EntityDataEntry
}

func (h *EntityDataDetachHistory) Redo() {
	defer tracing.NewRegion("EntityDataDetachHistory.Redo").End()
	h.Entity.DetachDataBinding(h.Data)
}

func (h *EntityDataDetachHistory) Undo() {
	defer tracing.NewRegion("EntityDataDetachHistory.Undo").End()
	h.Entity.AttachDataBinding(h.Data)
}

func (h *EntityDataDetachHistory) Delete() {}
func (h *EntityDataDetachHistory) Exit()   {}
