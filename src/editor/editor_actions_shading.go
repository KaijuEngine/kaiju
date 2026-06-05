/******************************************************************************/
/* editor_actions_shading.go                                                  */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace/shading_workspace"
	"kaijuengine.com/platform/hid"
)

func init() {
	registerEditorActionProvider(registerShadingActions)
}

func registerShadingActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          shading_workspace.ActionShadingShowCreateNodeMenu,
		Label:       "Show Create Node Menu",
		Description: "Opens the shader graph node creation menu.",
		Category:    "Shading",
		Tags:        []string{"shader", "graph", "node", "create", "menu"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  shading_workspace.ActionShadingShowCreateNodeMenu,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyC)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: shading_workspace.ID,
	}, ed.actionShadingShowCreateNodeMenu, ed.shadingCanRun)
	mustRegister(editor_action.Definition{
		ID:                shading_workspace.ActionShadingCreateNode,
		Label:             "Create Shader Node",
		Description:       "Creates a node in the shader graph.",
		Category:          "Shading",
		Tags:              []string{"shader", "graph", "node", "create"},
		DefaultParams:     editor_action.Params(shading_workspace.DefaultCreateNodeActionArgs()),
		NewParams:         func() any { return &shading_workspace.CreateNodeActionArgs{} },
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		Unbindable:        true,
		RequiredWorkspace: shading_workspace.ID,
		Variants:          shading_workspace.CreateNodeActionVariants(),
	}, ed.actionShadingCreateNode, ed.shadingCanRun)
}

func (ed *Editor) shadingCanRun(editor_action.Context, editor_action.Request) editor_action.Result {
	if _, ok := ed.shadingWorkspace(); !ok {
		return editor_action.Failure("shading workspace is not available")
	}
	return editor_action.Success("")
}

func (ed *Editor) actionShadingShowCreateNodeMenu(editor_action.Context, editor_action.Request) editor_action.Result {
	w, ok := ed.shadingWorkspace()
	if !ok {
		return editor_action.Failure("shading workspace is not available")
	}
	w.ShowCreateNodeMenu()
	return editor_action.Success("create node menu opened")
}

func (ed *Editor) actionShadingCreateNode(_ editor_action.Context, req editor_action.Request) editor_action.Result {
	w, ok := ed.shadingWorkspace()
	if !ok {
		return editor_action.Failure("shading workspace is not available")
	}
	args, ok := editor_action.Param[shading_workspace.CreateNodeActionArgs](req)
	if !ok {
		args = shading_workspace.DefaultCreateNodeActionArgs()
	}
	if _, ok = w.CreateNodeFromAction(args); !ok {
		return editor_action.Failure("failed to create shader node")
	}
	return editor_action.Success("shader node created")
}

func (ed *Editor) shadingWorkspace() (*shading_workspace.ShadingWorkspace, bool) {
	workspace, ok := ed.Workspace(shading_workspace.ID)
	if !ok {
		return nil, false
	}
	shading, ok := workspace.(*shading_workspace.ShadingWorkspace)
	return shading, ok
}
