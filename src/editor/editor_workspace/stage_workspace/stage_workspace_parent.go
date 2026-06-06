/******************************************************************************/
/* stage_workspace_parent.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/platform/profiler/tracing"
)

func (w *StageWorkspace) RemoveEntityFromParent(entity *editor_stage_manager.StageEntity) bool {
	defer tracing.NewRegion("StageWorkspace.RemoveEntityFromParent").End()
	if w == nil || w.stageView == nil ||
		entity == nil || entity.IsDeleted() || entity.IsLocked() || entity.Parent == nil {
		return false
	}
	w.stageView.Manager().SetEntityParent(entity, nil)
	return true
}

func (w *StageWorkspace) RemoveSelectedEntityFromParent() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.RemoveSelectedEntityFromParent").End()
	if w == nil || w.stageView == nil {
		return nil, false
	}
	selection := w.stageView.Manager().Selection()
	if len(selection) != 1 {
		return nil, false
	}
	entity := selection[0]
	return entity, w.RemoveEntityFromParent(entity)
}
