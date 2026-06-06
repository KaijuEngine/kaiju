/******************************************************************************/
/* history_transform.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type transformHistory struct {
	tman       *TransformationManager
	entities   []*editor_stage_manager.StageEntity
	toolTarget *editor_stage_manager.StageEntity
	from       []transformHistoryPRS
	to         []transformHistoryPRS
}

type transformHistoryPRS struct {
	position matrix.Vec3
	rotation matrix.Vec3
	scale    matrix.Vec3
}

func (h *transformHistory) Redo() {
	defer tracing.NewRegion("transformHistory.Redo").End()
	for i, e := range h.entities {
		e.Transform.SetPosition(h.to[i].position)
		e.Transform.SetRotation(h.to[i].rotation)
		e.Transform.SetScale(h.to[i].scale)
	}
	// TODO:  Use the following when the BVH.Refit function is fixed. Just
	// so there aren't any issues right now, I'm going to use a refit on
	// the first selected entity as it'll go to the root and refit all.
	// goroutine - Update all the BVHs
	//	for _, e := range t.stage.Manager().Selection() {
	//		e.StageData.Bvh.Refit()
	//	}
	man := h.tman.manager
	man.RefitBVH(h.entities[0])
	h.tman.translateTool.Show(h.toolTarget.Transform.Position())
}

func (h *transformHistory) Undo() {
	defer tracing.NewRegion("transformHistory.Undo").End()
	for i, e := range h.entities {
		e.Transform.SetPosition(h.from[i].position)
		e.Transform.SetRotation(h.from[i].rotation)
		e.Transform.SetScale(h.from[i].scale)
	}
	// TODO:  Use the following when the BVH.Refit function is fixed. Just
	// so there aren't any issues right now, I'm going to use a refit on
	// the first selected entity as it'll go to the root and refit all.
	// goroutine - Update all the BVHs
	//	for _, e := range t.stage.Manager().Selection() {
	//		e.StageData.Bvh.Refit()
	//	}
	man := h.tman.manager
	man.RefitBVH(h.entities[0])
	h.tman.translateTool.Show(h.toolTarget.Transform.Position())
}

func (h *transformHistory) Delete() {}
func (h *transformHistory) Exit()   {}
