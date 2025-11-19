package content_workspace

import (
	"kaiju/editor/editor_events"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
)

type ContentWorkspaceEditorInterface interface {
	Events() *editor_events.EditorEvents
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
	ShowReferences(id string)
}
