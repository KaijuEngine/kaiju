/******************************************************************************/
/* editor_actions.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"log/slog"

	"kaijuengine.com/editor/editor_action"
)

type editorActionRegistrar func(editor_action.Definition, editor_action.Handler, editor_action.CanRunFunc)
type editorActionProvider func(*Editor, editorActionRegistrar)

var editorActionProviders []editorActionProvider

func registerEditorActionProvider(provider editorActionProvider) {
	editorActionProviders = append(editorActionProviders, provider)
}

func (ed *Editor) initializeActions() {
	if ed.actions != nil {
		return
	}
	actions := editor_action.NewService()
	actions.SetContextProvider(ed.actionContext)
	actions.SetTransactionHooks(ed.history.BeginTransaction, ed.history.CommitTransaction, nil)
	if ed.host != nil {
		actions.SetMainThreadScheduler(ed.host.RunOnMainThread)
	}
	ed.actions = actions
	ed.registerBuiltinActions()
}

func (ed *Editor) actionContext() editor_action.Context {
	current := ""
	if ed.currentWorkspace != nil {
		current = ed.currentWorkspace.ID()
	}
	return editor_action.Context{
		CurrentWorkspace: current,
		InputFocused:     ed.IsInputFocused(),
		Services: map[string]any{
			"host":              ed.host,
			"history":           &ed.history,
			"settings":          &ed.settings,
			"project":           &ed.project,
			"projectFileSystem": ed.ProjectFileSystem(),
			"events":            &ed.events,
			"cache":             ed.Cache(),
			"stageView":         &ed.stageView,
		},
	}
}

func (ed *Editor) registerBuiltinActions() {
	for _, provider := range editorActionProviders {
		provider(ed, ed.mustRegisterAction)
	}
}

func (ed *Editor) mustRegisterAction(def editor_action.Definition, handler editor_action.Handler, canRun editor_action.CanRunFunc) {
	if err := ed.actions.Register(def, handler, canRun); err != nil {
		slog.Error("failed to register editor action", "id", def.ID, "error", err)
	}
}
