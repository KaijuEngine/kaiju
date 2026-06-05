/******************************************************************************/
/* editor_actions_render_graph.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_workspace/render_graph_workspace"
	"kaijuengine.com/platform/hid"
)

func init() {
	registerEditorActionProvider(registerRenderGraphActions)
}

func registerRenderGraphActions(ed *Editor, mustRegister editorActionRegistrar) {
	mustRegister(editor_action.Definition{
		ID:          render_graph_workspace.ActionRenderGraphShowCreateNodeMenu,
		Label:       "Show Create Node Menu",
		Description: "Opens the render graph node creation menu.",
		Category:    "Render Graph",
		Tags:        []string{"render", "graph", "node", "create", "menu"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  render_graph_workspace.ActionRenderGraphShowCreateNodeMenu,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyC)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: render_graph_workspace.ID,
	}, ed.actionRenderGraphShowCreateNodeMenu, ed.renderGraphCanRun)
	mustRegister(editor_action.Definition{
		ID:                render_graph_workspace.ActionRenderGraphCreateNode,
		Label:             "Create Render Graph Node",
		Description:       "Creates a node in the render graph.",
		Category:          "Render Graph",
		Tags:              []string{"render", "graph", "node", "create"},
		DefaultParams:     editor_action.Params(render_graph_workspace.DefaultCreateNodeActionArgs()),
		NewParams:         func() any { return &render_graph_workspace.CreateNodeActionArgs{} },
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		Unbindable:        true,
		RequiredWorkspace: render_graph_workspace.ID,
		Variants:          render_graph_workspace.CreateNodeActionVariants(),
	}, ed.actionRenderGraphCreateNode, ed.renderGraphCanRun)
}

func (ed *Editor) renderGraphCanRun(editor_action.Context, editor_action.Request) editor_action.Result {
	if _, ok := ed.renderGraphWorkspace(); !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	return editor_action.Success("")
}

func (ed *Editor) actionRenderGraphShowCreateNodeMenu(editor_action.Context, editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	w.ShowCreateNodeMenu()
	return editor_action.Success("create node menu opened")
}

func (ed *Editor) actionRenderGraphCreateNode(_ editor_action.Context, req editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	args, ok := editor_action.Param[render_graph_workspace.CreateNodeActionArgs](req)
	if !ok {
		args = render_graph_workspace.DefaultCreateNodeActionArgs()
	}
	if _, ok = w.CreateNodeFromAction(args); !ok {
		return editor_action.Failure("failed to create render graph node")
	}
	return editor_action.Success("render graph node created")
}

func (ed *Editor) renderGraphWorkspace() (*render_graph_workspace.RenderGraphWorkspace, bool) {
	workspace, ok := ed.Workspace(render_graph_workspace.ID)
	if !ok {
		return nil, false
	}
	renderGraph, ok := workspace.(*render_graph_workspace.RenderGraphWorkspace)
	return renderGraph, ok
}
