package editor_stage_manager

import "kaiju/platform/profiler/tracing"

type selectHistory struct {
	manager *StageManager
	from    []*StageEntity
	to      []*StageEntity
}

func (h *selectHistory) Redo() {
	defer tracing.NewRegion("selectHistory.Redo").End()
	prev := h.manager.selected
	h.manager.selected = h.to
	for _, e := range prev {
		h.manager.deselectEntityInternal(e)
	}
	for _, e := range h.manager.selected {
		h.manager.selectEntityInternal(e)
	}
}

func (h *selectHistory) Undo() {
	defer tracing.NewRegion("selectHistory.Undo").End()
	prev := h.manager.selected
	h.manager.selected = h.from
	for _, e := range prev {
		h.manager.deselectEntityInternal(e)
	}
	for _, e := range h.manager.selected {
		h.manager.selectEntityInternal(e)
	}
}

func (h *selectHistory) Delete() {}
func (h *selectHistory) Exit()   {}
