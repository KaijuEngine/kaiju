/******************************************************************************/
/* shader_designer_editor_interface.go                                        */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package shader_designer

import (
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
)

type ShaderDesignerEditorInterface interface {
	BlurInterface()
	FocusInterface()
	Events() *editor_events.EditorEvents
	ProjectFileSystem() *project_file_system.FileSystem
	Cache() *content_database.Cache
}
