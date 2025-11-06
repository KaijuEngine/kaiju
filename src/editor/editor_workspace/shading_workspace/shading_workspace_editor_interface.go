package shading_workspace

import (
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/memento"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
)

type ShadingWorkspaceEditorInterface interface {
	History() *memento.History
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
	StageView() *editor_stage_view.StageView
}
