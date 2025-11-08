package editor_stage_manager

import (
	"kaiju/platform/profiler/tracing"
)

type changeParentHistory struct {
	m          *StageManager
	e          *StageEntity
	prevParent *StageEntity
	nextParent *StageEntity
}

func (h *changeParentHistory) Redo() {
	defer tracing.NewRegion("changeParentHistory.Redo").End()
	h.m.SetEntityParent(h.e, h.nextParent)
}

func (h *changeParentHistory) Undo() {
	defer tracing.NewRegion("changeParentHistory.Undo").End()
	h.m.SetEntityParent(h.e, h.prevParent)
}

func (h *changeParentHistory) Delete() {}
func (h *changeParentHistory) Exit()   {}
