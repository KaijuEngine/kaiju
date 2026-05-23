/******************************************************************************/
/* constraint_authoring.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package stage_workspace

import (
	"weak"

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/platform/profiler/tracing"
)

func (w *StageWorkspace) ConnectSelectedAsDistanceChain() {
	w.connectSelectedAsConstraintChain(editor_stage_manager.ConstraintChainDistance)
}

func (w *StageWorkspace) ConnectSelectedAsRope() {
	w.connectSelectedAsConstraintChain(editor_stage_manager.ConstraintChainRope)
}

func (w *StageWorkspace) ConnectSelectedAsHingeChain() {
	w.connectSelectedAsConstraintChain(editor_stage_manager.ConstraintChainHinge)
}

func (w *StageWorkspace) connectSelectedAsConstraintChain(kind editor_stage_manager.ConstraintChainKind) {
	defer tracing.NewRegion("StageWorkspace.connectSelectedAsConstraintChain").End()
	man := w.stageView.Manager()
	w.ed.History().BeginTransaction()
	attachments := man.ConnectSelectedAsConstraintChain(kind)
	for _, attachment := range attachments {
		data_binding_renderer.Attached(attachment.Data, weak.Make(w.Host), man, attachment.Entity)
		data_binding_renderer.ShowSpecific(attachment.Data, weak.Make(w.Host), attachment.Entity)
		w.ed.History().Add(&constraintDataAttachHistory{
			workspace: w,
			Entity:    attachment.Entity,
			Data:      attachment.Data,
		})
	}
	if len(attachments) == 0 {
		w.ed.History().CancelTransaction()
		return
	}
	w.ed.History().CommitTransaction()
}
