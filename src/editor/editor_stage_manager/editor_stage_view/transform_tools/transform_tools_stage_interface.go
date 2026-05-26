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
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
)

type StageInterface interface {
	Camera() *editor_controls.EditorCamera
	WorkspaceHost() *engine.Host
	Manager() *editor_stage_manager.StageManager
	ViewportCursorPosition(mode editor_controls.EditorCameraMode, cursor *hid.Cursor) matrix.Vec2
	ViewportMousePosition(mouse *hid.Mouse) matrix.Vec2
	ViewportSize() matrix.Vec2
	ViewportReferenceSize() matrix.Vec2
}
