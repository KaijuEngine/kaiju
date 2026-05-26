/******************************************************************************/
/* history_stage_workspace_entity_lock.go                                     */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/platform/profiler/tracing"
)

type hierarchyEntityChangeLock struct {
	manager  *editor_stage_manager.StageManager
	entities []*editor_stage_manager.StageEntity
	previous []bool
	locked   bool
}

func (h *hierarchyEntityChangeLock) Redo() {
	defer tracing.NewRegion("hierarchyEntityChangeLock.Redo").End()
	for i := range h.entities {
		h.manager.SetEntityLocked(h.entities[i], h.locked)
	}
}

func (h *hierarchyEntityChangeLock) Undo() {
	defer tracing.NewRegion("hierarchyEntityChangeLock.Undo").End()
	for i := range h.entities {
		h.manager.SetEntityLocked(h.entities[i], h.previous[i])
	}
}

func (h *hierarchyEntityChangeLock) Delete() {}
func (h *hierarchyEntityChangeLock) Exit()   {}
