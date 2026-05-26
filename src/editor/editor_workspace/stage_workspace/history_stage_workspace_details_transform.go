/******************************************************************************/
/* history_stage_workspace_details_transform.go                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/matrix"
)

type transformHistoryType = int

const (
	transformHistoryTypePosition = transformHistoryType(iota)
	transformHistoryTypeRotation
	transformHistoryTypeScale
)

type detailTransformHistory struct {
	entities      []*editor_stage_manager.StageEntity
	transformType transformHistoryType
	prevValues    []matrix.Vec3
	nextValues    []matrix.Vec3
}

func (h *detailTransformHistory) Redo() {
	for i, e := range h.entities {
		switch h.transformType {
		case transformHistoryTypePosition:
			e.Transform.SetPosition(h.nextValues[i])
		case transformHistoryTypeRotation:
			e.Transform.SetRotation(h.nextValues[i])
		case transformHistoryTypeScale:
			e.Transform.SetScale(h.nextValues[i])
		}
	}
}

func (h *detailTransformHistory) Undo() {
	for i, e := range h.entities {
		switch h.transformType {
		case transformHistoryTypePosition:
			e.Transform.SetPosition(h.prevValues[i])
		case transformHistoryTypeRotation:
			e.Transform.SetRotation(h.prevValues[i])
		case transformHistoryTypeScale:
			e.Transform.SetScale(h.prevValues[i])
		}
	}
}

func (h *detailTransformHistory) Delete() {}
func (h *detailTransformHistory) Exit()   {}
