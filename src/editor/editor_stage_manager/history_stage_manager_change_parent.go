/******************************************************************************/
/* history_stage_manager_change_parent.go                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_manager

import (
	"kaijuengine.com/platform/profiler/tracing"
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
