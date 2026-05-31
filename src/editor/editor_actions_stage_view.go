/******************************************************************************/
/* editor_actions_stage_view.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace/stage_workspace"
	"kaijuengine.com/platform/hid"
)

const (
	ActionStageSetGridVisible       editor_action.ActionID = "stage.setGridVisible"
	ActionStageToggleViewportLayout editor_action.ActionID = "stage.toggleViewportLayout"
	ActionStageToggleContentPanel   editor_action.ActionID = "stage.toggleContentPanel"
	ActionStageToggleHierarchyPanel editor_action.ActionID = "stage.toggleHierarchyPanel"
	ActionStageToggleDetailsPanel   editor_action.ActionID = "stage.toggleDetailsPanel"
	ActionStageRenameActor          editor_action.ActionID = "stage.renameActor"
	ActionStageFocusSelection       editor_action.ActionID = "stage.focusSelection"
)

type gridVisibleActionArgs struct {
	Visible bool `json:"visible"`
}

func init() {
	registerEditorActionProvider(registerStageViewActions)
}

func registerStageViewActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          ActionStageToggleViewportLayout,
		Label:       "Split/Focus Viewport",
		Description: "Toggles the stage between one focused viewport and the split viewport layout.",
		Category:    "Stage",
		Tags:        []string{"viewport", "split", "focus", "layout"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageToggleViewportLayout,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyP)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionToggleViewportLayout, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageToggleContentPanel,
		Label:       "Toggle Content Panel",
		Description: "Shows or hides the stage content panel.",
		Category:    "Stage",
		Tags:        []string{"content", "panel", "visibility"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageToggleContentPanel,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyC)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionToggleContentPanel, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageToggleHierarchyPanel,
		Label:       "Toggle Hierarchy Panel",
		Description: "Shows or hides the stage hierarchy panel.",
		Category:    "Stage",
		Tags:        []string{"hierarchy", "panel", "visibility"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageToggleHierarchyPanel,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyH)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionToggleHierarchyPanel, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageToggleDetailsPanel,
		Label:       "Toggle Details Panel",
		Description: "Shows or hides the stage details panel.",
		Category:    "Stage",
		Tags:        []string{"details", "panel", "visibility"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageToggleDetailsPanel,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyD)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionToggleDetailsPanel, ed.stageCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageRenameActor,
		Label:       "Rename Actor",
		Description: "Focuses the selected stage actor name field.",
		Category:    "Stage",
		Tags:        []string{"actor", "entity", "rename", "name"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageRenameActor,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF2)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionRenameActor, ed.stageSelectionCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageFocusSelection,
		Label:       "Focus Selection",
		Description: "Frames the selected stage actor in the viewport.",
		Category:    "Stage",
		Tags:        []string{"actor", "entity", "selection", "focus", "frame", "viewport"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  ActionStageFocusSelection,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: stage_workspace.ID,
	}, ed.actionFocusSelection, ed.stageSelectionCanRun)
	mustRegister(editor_action.Definition{
		ID:          ActionStageSetGridVisible,
		Label:       "Set Grid Visible",
		Description: "Shows or hides the stage viewport grid.",
		Category:    "Stage",
		Tags:        []string{"grid", "view", "visibility"},
		Parameters:  []editor_action.Parameter{{Name: "visible", Label: "Visible", Type: "bool"}},
		NewParams:   func() any { return &gridVisibleActionArgs{} },
		UndoPolicy:  editor_action.UndoPolicyNone,
		Visible:     false,
		Unbindable:  true,
	}, ed.actionSetGridVisible, nil)
}

func (ed *Editor) actionToggleViewportLayout(editor_action.Context, editor_action.Request) editor_action.Result {
	if !ed.StageWorkspace().ToggleViewportSplitFocus() {
		return editor_action.Failure("stage viewport layout was not changed")
	}
	return editor_action.Success("stage viewport layout changed")
}

func (ed *Editor) actionToggleContentPanel(editor_action.Context, editor_action.Request) editor_action.Result {
	if !ed.StageWorkspace().ToggleContentPanel() {
		return editor_action.Failure("stage content panel was not changed")
	}
	return editor_action.Success("stage content panel changed")
}

func (ed *Editor) actionToggleHierarchyPanel(editor_action.Context, editor_action.Request) editor_action.Result {
	if !ed.StageWorkspace().ToggleHierarchyPanel() {
		return editor_action.Failure("stage hierarchy panel was not changed")
	}
	return editor_action.Success("stage hierarchy panel changed")
}

func (ed *Editor) actionToggleDetailsPanel(editor_action.Context, editor_action.Request) editor_action.Result {
	if !ed.StageWorkspace().ToggleDetailsPanel() {
		return editor_action.Failure("stage details panel was not changed")
	}
	return editor_action.Success("stage details panel changed")
}

func (ed *Editor) actionRenameActor(editor_action.Context, editor_action.Request) editor_action.Result {
	if !ed.StageWorkspace().FocusRename() {
		return editor_action.Failure("stage actor rename was not focused")
	}
	return stageSelectionResult("stage actor rename focused", ed.stageView.Manager().Selection())
}

func (ed *Editor) actionFocusSelection(editor_action.Context, editor_action.Request) editor_action.Result {
	if !ed.stageView.FocusSelection() {
		return editor_action.Failure("stage selection was not focused")
	}
	return stageSelectionResult("stage selection focused", ed.stageView.Manager().Selection())
}

func (ed *Editor) actionSetGridVisible(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args, ok := editor_action.Param[gridVisibleActionArgs](req)
	if !ok {
		return editor_action.Failure("visible is required")
	}
	ed.SetGridVisible(args.Visible)
	return editor_action.Success("grid visibility changed")
}
