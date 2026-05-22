/******************************************************************************/
/* editor_stage_view_workspace.go                                             */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
)

type EditorStageViewWorkspaceInterface interface {
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
	FocusInterface()
	BlurInterface()
	Settings() *editor_settings.Settings
	StageView() *StageView
}
