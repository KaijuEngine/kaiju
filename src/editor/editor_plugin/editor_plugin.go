/******************************************************************************/
/* editor_plugin.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_plugin

import (
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

type EditorPlugin interface {
	Launch(EditorInterface) error
}

// EditorInterface is what plugins receive in their Launch() callback. It is
// a superset of editor_workspace.WorkspaceEditorInterface — anything a
// built-in workspace can do, a plugin can do during runtime.
//
// To register a workspace, plugins call
// editor_workspace_registry.Register(&MyWorkspace{}) from their package
// init() (alongside the existing editor.RegisterPlugin call). The editor's
// reconcile step picks up late registrations on plugin launch and rebuilds
// the menu bar tab strip.
type EditorInterface interface {
	Host() *engine.Host
	BlurInterface()
	FocusInterface()
	Settings() *editor_settings.Settings
	Events() *editor_events.EditorEvents
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	StageView() *editor_stage_view.StageView

	// Workspace registry access.
	SelectWorkspace(id string) error
	Workspace(id string) (editor_workspace.Workspace, bool)
	Workspaces() []editor_workspace.Workspace
}
