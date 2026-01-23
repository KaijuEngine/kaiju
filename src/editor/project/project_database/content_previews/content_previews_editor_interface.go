package content_previews

import (
	"kaiju/editor/editor_events"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
)

type EditorInterface interface {
	Host() *engine.Host
	Events() *editor_events.EditorEvents
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
}
