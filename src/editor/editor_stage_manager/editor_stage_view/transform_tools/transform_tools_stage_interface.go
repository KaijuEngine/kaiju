/******************************************************************************/
/* transform_tools_stage_interface.go                                         */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package transform_tools

import (
	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/engine"
)

type StageInterface interface {
	Camera() *editor_controls.EditorCamera
	WorkspaceHost() *engine.Host
	Manager() *editor_stage_manager.StageManager
}
