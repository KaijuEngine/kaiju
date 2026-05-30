/******************************************************************************/
/* workspace_editor_interface.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_workspace

import (
	"kaijuengine.com/editor/editor_action"
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

// WorkspaceEditorInterface is the single editor surface every workspace
// receives during Initialize. It intentionally exposes editor-level services
// (host, settings, events, history, project, stage view, content previewer)
// plus the workspace registry and switching API, but does not contain any
// per-workspace methods. Cross-workspace operations go through events
// (Events()) or through Workspace(id) lookups against well-known string IDs
// or typed service interfaces.
//
// Methods on this interface map 1:1 to methods on the Editor struct so the
// editor implements the interface implicitly.
type WorkspaceEditorInterface interface {
	// Engine / runtime
	Host() *engine.Host
	Cache() *content_database.Cache
	ContentPreviewer() *content_previews.ContentPreviewer

	// Editor services
	Actions() *editor_action.Service
	Settings() *editor_settings.Settings
	Events() *editor_events.EditorEvents
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	StageView() *editor_stage_view.StageView

	// Focus management — workspaces blur the rest of the editor while a
	// modal/overlay is in front of them and re-focus on close.
	BlurInterface()
	FocusInterface()
	IsInputFocused() bool

	// Workspace registry. SelectWorkspace switches the active workspace to
	// the one with the given id (no-op if unknown or disabled). Workspace
	// returns the live instance for type-asserted typed-service queries.
	// Workspaces returns the enabled set in current load order.
	SelectWorkspace(id string) error
	Workspace(id string) (Workspace, bool)
	Workspaces() []Workspace

	// UpdateSettings persists the current Settings struct and re-applies
	// frame rate / scroll speed / etc. to the live host.
	UpdateSettings()

	// ShowReferences opens the references viewer overlay for the given
	// content id. Lives here because the overlay is editor-owned, not
	// workspace-owned.
	ShowReferences(id string)
}
