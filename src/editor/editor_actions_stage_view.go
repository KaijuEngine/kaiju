/******************************************************************************/
/* editor_actions_stage_view.go                                               */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import "kaijuengine.com/editor/editor_action"

const (
	ActionStageSetGridVisible editor_action.ActionID = "stage.setGridVisible"
)

type gridVisibleActionArgs struct {
	Visible bool `json:"visible"`
}

func init() {
	registerEditorActionProvider(registerStageViewActions)
}

func registerStageViewActions(ed *Editor, mustRegister editorActionRegistrar) {
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
	}, ed.actionSetGridVisible, nil)
}

func (ed *Editor) actionSetGridVisible(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args, ok := editor_action.Param[gridVisibleActionArgs](req)
	if !ok {
		return editor_action.Failure("visible is required")
	}
	ed.SetGridVisible(args.Visible)
	return editor_action.Success("grid visibility changed")
}
