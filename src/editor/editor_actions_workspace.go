/******************************************************************************/
/* editor_actions_workspace.go                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"fmt"
	"strings"

	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace_registry"
)

const (
	ActionEditorOpenWorkspace editor_action.ActionID = "editor.openWorkspace"
)

type workspaceActionArgs struct {
	Workspace string `json:"workspace"`
}

func init() {
	registerEditorActionProvider(registerWorkspaceActions)
}

func registerWorkspaceActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          ActionEditorOpenWorkspace,
		Label:       "Open Workspace",
		Description: "Switches to another editor workspace.",
		Category:    "Workspace",
		Tags:        []string{"workspace", "tab"},
		Parameters: []editor_action.Parameter{
			{Name: "workspace", Label: "Workspace", Type: "string", Required: true},
		},
		NewParams:  func() any { return &workspaceActionArgs{} },
		UndoPolicy: editor_action.UndoPolicyNone,
		Visible:    false,
		Unbindable: true,
	}, ed.actionOpenWorkspace, nil)
	ed.registerRegisteredWorkspaceActions()
}

func (ed *Editor) registerRegisteredWorkspaceActions() {
	if ed.actions == nil {
		return
	}
	for _, workspace := range editor_workspace_registry.All() {
		id := strings.TrimSpace(workspace.ID())
		if id == "" {
			continue
		}
		actionID := openWorkspaceActionID(id)
		if _, ok := ed.actions.Registry().Definition(actionID); ok {
			continue
		}
		displayName := strings.TrimSpace(workspace.DisplayName())
		if displayName == "" {
			displayName = id
		}
		params := workspaceActionArgs{Workspace: id}
		ed.mustRegisterAction(editor_action.Definition{
			ID:            actionID,
			Label:         fmt.Sprintf("Open %s Workspace", displayName),
			Description:   fmt.Sprintf("Switches to the %s workspace.", displayName),
			Category:      "Workspace",
			Tags:          []string{"workspace", "tab", id, displayName},
			DefaultParams: editor_action.Params(params),
			NewParams:     func() any { return &workspaceActionArgs{} },
			UndoPolicy:    editor_action.UndoPolicyNone,
			Visible:       true,
		}, ed.actionOpenWorkspace, ed.workspaceCanRun(id))
	}
}

func openWorkspaceActionID(workspaceID string) editor_action.ActionID {
	return editor_action.ActionID("editor.openWorkspace." + strings.TrimSpace(workspaceID))
}

func (ed *Editor) workspaceCanRun(workspaceID string) editor_action.CanRunFunc {
	return func(editor_action.Context, editor_action.Request) editor_action.Result {
		if _, ok := ed.Workspace(workspaceID); !ok {
			return editor_action.Failure("workspace is not available")
		}
		return editor_action.Success("")
	}
}

func (ed *Editor) actionOpenWorkspace(ctx editor_action.Context, req editor_action.Request) editor_action.Result {
	args, ok := editor_action.Param[workspaceActionArgs](req)
	if !ok {
		return editor_action.Failure("workspace is required")
	}
	args.Workspace = strings.TrimSpace(args.Workspace)
	if args.Workspace == "" {
		return editor_action.Failure("workspace is required")
	}
	if err := ed.SelectWorkspace(args.Workspace); err != nil {
		return editor_action.Failure(err.Error())
	}
	return editor_action.Success("workspace opened")
}
