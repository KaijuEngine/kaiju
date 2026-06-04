/******************************************************************************/
/* stage_workspace_align.go                                                   */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"weak"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/profiler/tracing"
)

func (w *StageWorkspace) AlignEntityWithView(entity *editor_stage_manager.StageEntity) bool {
	defer tracing.NewRegion("StageWorkspace.AlignEntityWithView").End()
	if w == nil || w.Host == nil || entity == nil || entity.IsDeleted() || entity.IsLocked() {
		return false
	}
	cam := w.Host.PrimaryCamera()
	position := cam.Position()
	rotation := viewAlignedRotation(cam.Up(), cam.Forward())
	posHistory := &detailTransformHistory{
		entities:      []*editor_stage_manager.StageEntity{entity},
		transformType: transformHistoryTypePosition,
		prevValues:    []matrix.Vec3{entity.Transform.Position()},
	}
	rotHistory := &detailTransformHistory{
		entities:      []*editor_stage_manager.StageEntity{entity},
		transformType: transformHistoryTypeRotation,
		prevValues:    []matrix.Vec3{entity.Transform.Rotation()},
	}
	history := w.ed.History()
	history.BeginTransaction()
	defer history.CommitTransaction()
	entity.Transform.SetWorldPosition(position)
	entity.Transform.SetWorldRotation(rotation)
	posHistory.nextValues = []matrix.Vec3{entity.Transform.Position()}
	rotHistory.nextValues = []matrix.Vec3{entity.Transform.Rotation()}
	history.Add(posHistory)
	history.Add(rotHistory)
	w.stageView.Manager().RefitBVH(entity)
	for _, db := range entity.DataBindings() {
		data_binding_renderer.Updated(db, weak.Make(w.Host), entity)
	}
	return true
}

func (w *StageWorkspace) AlignSelectedEntityWithView() (*editor_stage_manager.StageEntity, bool) {
	defer tracing.NewRegion("StageWorkspace.AlignSelectedEntityWithView").End()
	if w == nil || w.stageView == nil {
		return nil, false
	}
	selection := w.stageView.Manager().Selection()
	if len(selection) != 1 {
		return nil, false
	}
	entity := selection[0]
	return entity, w.AlignEntityWithView(entity)
}

func viewAlignedRotation(up, forward matrix.Vec3) matrix.Vec3 {
	forward = forward.Normal()
	up = up.Normal()
	right := matrix.Vec3Cross(up, forward).Normal()
	up = matrix.Vec3Cross(forward, right).Normal()
	rot := matrix.Mat4{
		right.X(), right.Y(), right.Z(), 0,
		up.X(), up.Y(), up.Z(), 0,
		forward.X(), forward.Y(), forward.Z(), 0,
		0, 0, 0, 1,
	}
	return rot.ExtractRotation().ToEuler()
}
