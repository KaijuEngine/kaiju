/******************************************************************************/
/* editor_actions_stage_selection.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/platform/hid"
)

const (
	ActionStageSelectAll        editor_action.ActionID = "stage.selectAll"
	ActionStageClearSelection   editor_action.ActionID = "stage.clearSelection"
	ActionStageSelectByID       editor_action.ActionID = "stage.selectById"
	ActionStageDeleteSelection  editor_action.ActionID = "stage.deleteSelection"
	ActionStageDuplicate        editor_action.ActionID = "stage.duplicateSelection"
	ActionStageRemoveFromParent editor_action.ActionID = "stage.removeFromParent"
)

type selectByIDActionArgs struct {
	ID string `json:"id"`
}

func init() {
	registerEditorActionProvider(registerStageSelectionActions)
}

func registerStageSelectionActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:                ActionStageSelectAll,
		Label:             "Select All",
		Description:       "Selects every unlocked entity in the current stage.",
		Category:          "Stage",
		Tags:              []string{"selection", "all"},
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionSelectAll, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:                ActionStageClearSelection,
		Label:             "Clear Selection",
		Description:       "Deselects every selected stage entity.",
		Category:          "Stage",
		Tags:              []string{"selection", "deselect"},
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionClearSelection, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:                ActionStageSelectByID,
		Label:             "Select Entity By ID",
		Description:       "Selects one entity by its stable id.",
		Category:          "Stage",
		Tags:              []string{"selection", "entity"},
		Parameters:        []editor_action.Parameter{{Name: "id", Label: "Entity ID", Type: "string", Required: true}},
		NewParams:         func() any { return &selectByIDActionArgs{} },
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           false,
		Unbindable:        true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionSelectByID, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageDeleteSelection,
		Label:       "Delete Selection",
		Description: "Deletes selected stage entities.",
		Category:    "Stage",
		Tags:        []string{"delete", "remove", "selection"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageDeleteSelection,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyDelete)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionDeleteSelection, ed.stageSelectionCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageDuplicate,
		Label:       "Duplicate Selection",
		Description: "Duplicates selected stage entities.",
		Category:    "Stage",
		Tags:        []string{"duplicate", "copy", "selection"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageDuplicate,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyD)}, CtrlOrMeta: true},
		}},
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionDuplicateSelection, ed.stageSelectionCanRun)
	mustRegister(editor_action.Definition{
		ID:                ActionStageRemoveFromParent,
		Label:             "Remove from parent",
		Description:       "Moves the selected stage entity to the root of the hierarchy.",
		Category:          "Stage",
		Tags:              []string{"actor", "entity", "selection", "hierarchy", "parent", "root", "remove"},
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionRemoveFromParent, ed.stageSingleSelectionWithParentCanRun)
}

func (ed *Editor) actionSelectAll(editor_action.Context, editor_action.Request) editor_action.Result {
	man := ed.stageView.Manager()
	man.SelectEntities(man.List(), editor_stage_manager.SelectionModeReplace)
	return stageSelectionResult("all entities selected", man.Selection())
}

func (ed *Editor) actionClearSelection(editor_action.Context, editor_action.Request) editor_action.Result {
	man := ed.stageView.Manager()
	man.ClearSelection()
	return stageSelectionResult("selection cleared", man.Selection())
}

func (ed *Editor) actionSelectByID(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args, ok := editor_action.Param[selectByIDActionArgs](req)
	if !ok {
		return editor_action.Failure("id is required")
	}
	args.ID = strings.TrimSpace(args.ID)
	if args.ID == "" {
		return editor_action.Failure("id is required")
	}
	man := ed.stageView.Manager()
	if _, ok := man.EntityById(args.ID); !ok {
		return editor_action.Failure("entity was not found")
	}
	man.SelectEntityById(args.ID)
	return stageSelectionResult("entity selected", man.Selection())
}

func (ed *Editor) actionDeleteSelection(editor_action.Context, editor_action.Request) editor_action.Result {
	man := ed.stageView.Manager()
	before := stageEntityIDs(man.Selection())
	man.DestroySelected()
	result := stageSelectionResult("selection deleted", man.Selection())
	result.AffectedEntityIDs = before
	return result
}

func (ed *Editor) actionDuplicateSelection(editor_action.Context, editor_action.Request) editor_action.Result {
	ed.stageView.DuplicateSelected(&ed.project)
	return stageSelectionResult("selection duplicated", ed.stageView.Manager().Selection())
}

func (ed *Editor) actionRemoveFromParent(editor_action.Context, editor_action.Request) editor_action.Result {
	entity, ok := ed.StageWorkspace().RemoveSelectedEntityFromParent()
	if !ok {
		return editor_action.Failure("selected entity was not removed from parent")
	}
	return stageResult("selected entity removed from parent", entity, ed.stageView.Manager().Selection())
}
