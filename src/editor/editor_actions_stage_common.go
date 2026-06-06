/******************************************************************************/
/* editor_actions_stage_common.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_workspace/stage_workspace"
)

func (ed *Editor) stageCanRun(editor_action.Context, editor_action.Request) editor_action.Result {
	if _, ok := ed.Workspace(stage_workspace.ID); !ok {
		return editor_action.Failure("stage workspace is not available")
	}
	return editor_action.Success("")
}

func (ed *Editor) stageSelectionCanRun(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	if can := ed.stageCanRun(ctx, req); !can.OK {
		return can
	}
	if !ed.stageView.Manager().HasSelection() {
		return editor_action.Failure("nothing is selected")
	}
	return editor_action.Success("")
}

func (ed *Editor) stageSingleSelectionCanRun(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	if can := ed.stageCanRun(ctx, req); !can.OK {
		return can
	}
	if len(ed.stageView.Manager().Selection()) != 1 {
		return editor_action.Failure("exactly one entity must be selected")
	}
	return editor_action.Success("")
}

func (ed *Editor) stageSingleSelectionWithParentCanRun(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	if can := ed.stageSingleSelectionCanRun(ctx, req); !can.OK {
		return can
	}
	entity := ed.stageView.Manager().Selection()[0]
	if entity == nil || entity.IsDeleted() {
		return editor_action.Failure("selected entity is not available")
	}
	if entity.IsLocked() {
		return editor_action.Failure("selected entity is locked")
	}
	if entity.Parent == nil {
		return editor_action.Failure("selected entity has no parent")
	}
	return editor_action.Success("")
}

func (ed *Editor) stageTransformToolCanRun(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	if can := ed.stageSelectionCanRun(ctx, req); !can.OK {
		return can
	}
	if req.Source == editor_action.SourceKeybind && !ed.stageView.CanUseTransformToolKeybinding() {
		return editor_action.Failure("stage viewport is not ready for transform tool hotkeys")
	}
	return editor_action.Success("")
}

func stageResult(message string, affected *editor_stage_manager.StageEntity, selected []*editor_stage_manager.StageEntity) editor_action.Result {
	result := stageSelectionResult(message, selected)
	if affected != nil {
		result.AffectedEntityIDs = []string{affected.StageData.Description.Id}
	}
	return result
}

func stageSelectionResult(message string, selected []*editor_stage_manager.StageEntity) editor_action.Result {
	result := editor_action.Success(message)
	result.SelectedEntityIDs = stageEntityIDs(selected)
	return result
}

func stageEntityIDs(entities []*editor_stage_manager.StageEntity) []string {
	ids := make([]string, 0, len(entities))
	for _, e := range entities {
		if e == nil {
			continue
		}
		ids = append(ids, e.StageData.Description.Id)
	}
	return ids
}
