package stage_workspace

import (
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/project/project_file_system"
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
