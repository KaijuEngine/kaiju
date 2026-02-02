package stage_workspace

import (
	"kaiju/editor/editor_events"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_database/content_previews"
	"kaiju/editor/project/project_file_system"
)

type StageWorkspaceEditorInterface interface {
	Events() *editor_events.EditorEvents
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
	FocusInterface()
	BlurInterface()
	Settings() *editor_settings.Settings
	StageView() *editor_stage_view.StageView
	ShowReferences(id string)
	ContentWorkspaceSelected()
	ContentPreviewer() *content_previews.ContentPreviewer
}
