package stage_workspace

import (
	"kaiju/engine"
	"kaiju/games/editor/editor_workspace/common_workspace"
)

type Workspace struct {
	common_workspace.CommonWorkspace
}

func (w *Workspace) Initialize(host *engine.Host) {
	w.CommonWorkspace.Initialize(host, "editor/ui/workspace/stage_workspace.go.html", nil, nil)
}
