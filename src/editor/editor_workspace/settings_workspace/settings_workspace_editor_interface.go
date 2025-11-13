package settings_workspace

import (
	"kaiju/editor/editor_settings"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/editor/project/project_file_system"
)

type SettingsWorkspaceEditorInterface interface {
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	Settings() *editor_settings.Settings
}
