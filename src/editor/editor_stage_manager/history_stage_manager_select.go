/******************************************************************************/
/* stage_manager_select_history.go                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"slices"

	"kaijuengine.com/platform/profiler/tracing"
)

type selectHistory struct {
	manager *StageManager
	from    []*StageEntity
	to      []*StageEntity
}

func (h *selectHistory) Redo() {
	defer tracing.NewRegion("selectHistory.Redo").End()
	for i := len(h.manager.selected) - 1; i >= 0; i-- {
		if !slices.Contains(h.to, h.manager.selected[i]) {
			h.manager.DeselectEntity(h.manager.selected[i])
		}
	}
	for _, e := range h.to {
		h.manager.SelectEntity(e)
	}
}

func (h *selectHistory) Undo() {
	defer tracing.NewRegion("selectHistory.Undo").End()
	for i := len(h.manager.selected) - 1; i >= 0; i-- {
		if !slices.Contains(h.from, h.manager.selected[i]) {
			h.manager.DeselectEntity(h.manager.selected[i])
		}
	}
	for _, e := range h.from {
		h.manager.SelectEntity(e)
	}
}

func (h *selectHistory) Delete() {}
func (h *selectHistory) Exit()   {}
