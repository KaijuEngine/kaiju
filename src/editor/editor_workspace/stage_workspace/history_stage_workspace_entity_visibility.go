/******************************************************************************/
/* history_stage_workspace_entity_visibility.go                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/platform/profiler/tracing"
)

type hierarchyEntityChangeVisibilty struct {
	entities []*editor_stage_manager.StageEntity
	previous []bool
	visible  bool
}

func (h *hierarchyEntityChangeVisibilty) Redo() {
	defer tracing.NewRegion("hierarchyEntityChangeVisibilty.Redo").End()
	for i := range h.entities {
		h.entities[i].SetActive(h.visible)
	}
}

func (h *hierarchyEntityChangeVisibilty) Undo() {
	defer tracing.NewRegion("hierarchyEntityChangeVisibilty.Undo").End()
	for i := range h.entities {
		h.entities[i].SetActive(h.previous[i])
	}
}

func (h *hierarchyEntityChangeVisibilty) Delete() {}
func (h *hierarchyEntityChangeVisibilty) Exit()   {}
