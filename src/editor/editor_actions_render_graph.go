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
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyN)}},
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
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		Unbindable:        true,
		RequiredWorkspace: render_graph_workspace.ID,
		Variants:          render_graph_workspace.CreateNodeActionVariants(),
	}, ed.actionRenderGraphCreateNode, ed.renderGraphCanRun)
	mustRegister(editor_action.Definition{
		ID:                render_graph_workspace.ActionRenderGraphCreateComment,
		Label:             "Create Render Graph Comment",
		Description:       "Creates a resizable comment block in the render graph.",
		Category:          "Render Graph",
		Tags:              []string{"render", "graph", "comment", "block", "note", "create"},
		DefaultParams:     editor_action.Params(render_graph_workspace.DefaultCreateCommentActionArgs()),
		NewParams:         func() any { return &render_graph_workspace.CreateCommentActionArgs{} },
		UndoPolicy:        editor_action.UndoPolicyTransaction,
		Visible:           true,
		Unbindable:        true,
		RequiredWorkspace: render_graph_workspace.ID,
	}, ed.actionRenderGraphCreateComment, ed.renderGraphCanRun)
	mustRegister(editor_action.Definition{
		ID:          render_graph_workspace.ActionRenderGraphCenterView,
		Label:       "Center Render Graph View",
		Description: "Returns the render graph view to the center of the graph area.",
		Category:    "Render Graph",
		Tags:        []string{"render", "graph", "view", "center", "pan"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  render_graph_workspace.ActionRenderGraphCenterView,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKey0)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: render_graph_workspace.ID,
	}, ed.actionRenderGraphCenterView, ed.renderGraphCanRun)
	mustRegister(editor_action.Definition{
		ID:          render_graph_workspace.ActionRenderGraphFocusSelection,
		Label:       "Focus Render Graph Selection",
		Description: "Centers the render graph view on the selected nodes.",
		Category:    "Render Graph",
		Tags:        []string{"render", "graph", "node", "selection", "focus", "frame", "view"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  render_graph_workspace.ActionRenderGraphFocusSelection,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyF)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: render_graph_workspace.ID,
	}, ed.actionRenderGraphFocusSelection, ed.renderGraphCanRun)
	mustRegister(editor_action.Definition{
		ID:          render_graph_workspace.ActionRenderGraphSave,
		Label:       "Save Render Graph",
		Description: "Saves the current render graph.",
		Category:    "Render Graph",
		Tags:        []string{"render", "graph", "save", "content"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  render_graph_workspace.ActionRenderGraphSave,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyS)}, CtrlOrMeta: true},
		}},
		UndoPolicy:        editor_action.UndoPolicyNone,
		Visible:           true,
		RequiredWorkspace: render_graph_workspace.ID,
	}, ed.actionRenderGraphSave, ed.renderGraphCanRun)
	mustRegister(editor_action.Definition{
		ID:          render_graph_workspace.ActionRenderGraphDeleteSelection,
		Label:       "Delete Render Graph Selection",
		Description: "Deletes the selected render graph nodes.",
		Category:    "Render Graph",
		Tags:        []string{"render", "graph", "node", "selection", "delete", "remove"},
		DefaultBindings: []editor_action.ActionBinding{{
			Action:  render_graph_workspace.ActionRenderGraphDeleteSelection,
			Enabled: true,
			Chord:   editor_action.KeyChord{Keys: []int{int(hid.KeyboardKeyDelete)}},
		}},
		UndoPolicy:        editor_action.UndoPolicyManaged,
		Visible:           true,
		RequiredWorkspace: render_graph_workspace.ID,
	}, ed.actionRenderGraphDeleteSelection, ed.renderGraphCanRun)
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

func (ed *Editor) actionRenderGraphCreateComment(_ editor_action.Context, req editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	args, ok := editor_action.Param[render_graph_workspace.CreateCommentActionArgs](req)
	if !ok {
		args = render_graph_workspace.DefaultCreateCommentActionArgs()
	}
	if _, ok = w.CreateCommentFromAction(args); !ok {
		return editor_action.Failure("failed to create render graph comment")
	}
	return editor_action.Success("render graph comment created")
}

func (ed *Editor) actionRenderGraphCenterView(editor_action.Context, editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	w.CenterView()
	return editor_action.Success("render graph view centered")
}

func (ed *Editor) actionRenderGraphFocusSelection(editor_action.Context, editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	if !w.FocusSelectedNodes() {
		return editor_action.Failure("no render graph nodes are selected")
	}
	return editor_action.Success("render graph selection focused")
}

func (ed *Editor) actionRenderGraphSave(editor_action.Context, editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	w.SaveCurrentGraph()
	return editor_action.Success("render graph saved")
}

func (ed *Editor) actionRenderGraphDeleteSelection(editor_action.Context, editor_action.Request) editor_action.Result {
	w, ok := ed.renderGraphWorkspace()
	if !ok {
		return editor_action.Failure("render graph workspace is not available")
	}
	if !w.DeleteSelectedNodes() {
		return editor_action.Failure("no render graph nodes are selected")
	}
	return editor_action.Success("render graph selection deleted")
}

func (ed *Editor) renderGraphWorkspace() (*render_graph_workspace.RenderGraphWorkspace, bool) {
	workspace, ok := ed.Workspace(render_graph_workspace.ID)
	if !ok {
		return nil, false
	}
	renderGraph, ok := workspace.(*render_graph_workspace.RenderGraphWorkspace)
	return renderGraph, ok
}
