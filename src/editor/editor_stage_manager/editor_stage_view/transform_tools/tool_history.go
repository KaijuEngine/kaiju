/******************************************************************************/
/* tool_history.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

type toolHistory struct {
	stage    StageInterface
	entities []*editor_stage_manager.StageEntity
	from     []matrix.Vec3
	to       []matrix.Vec3
	state    ToolState
}

func (h *toolHistory) Redo() {
	defer tracing.NewRegion("toolHistory.Redo").End()
	for i, e := range h.entities {
		switch h.state {
		case ToolStateMove:
			e.Transform.SetPosition(h.to[i])
		case ToolStateRotate:
			e.Transform.SetRotation(h.to[i])
		case ToolStateScale:
			e.Transform.SetScale(h.to[i])
		}
	}
	// TODO:  Use the following when the BVH.Refit function is fixed. Just
	// so there aren't any issues right now, I'm going to use a refit on
	// the first selected entity as it'll go to the root and refit all.
	// goroutine - Update all the BVHs
	//	for _, e := range t.stage.Manager().Selection() {
	//		e.StageData.Bvh.Refit()
	//	}
	h.stage.Manager().RefitBVH(h.entities[0])
}

func (h *toolHistory) Undo() {
	defer tracing.NewRegion("toolHistory.Undo").End()
	for i, e := range h.entities {
		switch h.state {
		case ToolStateMove:
			e.Transform.SetPosition(h.from[i])
		case ToolStateRotate:
			e.Transform.SetRotation(h.from[i])
		case ToolStateScale:
			e.Transform.SetScale(h.from[i])
		}
	}
	// TODO:  Use the following when the BVH.Refit function is fixed. Just
	// so there aren't any issues right now, I'm going to use a refit on
	// the first selected entity as it'll go to the root and refit all.
	// goroutine - Update all the BVHs
	//	for _, e := range t.stage.Manager().Selection() {
	//		e.StageData.Bvh.Refit()
	//	}
	h.stage.Manager().RefitBVH(h.entities[0])
}

func (h *toolHistory) Delete() {}
func (h *toolHistory) Exit()   {}
