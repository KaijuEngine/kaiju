package transform_tools

import (
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_stage_manager"
	"kaiju/engine"
)

type StageInterface interface {
	Camera() *editor_controls.EditorCamera
	WorkspaceHost() *engine.Host
	Manager() *editor_stage_manager.StageManager
}
