package content_previews

import (
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

type EditorInterface interface {
	Host() *engine.Host
	Events() *editor_events.EditorEvents
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
}
