package content_workspace

import (
	"kaiju/engine"
	"kaiju/games/editor/editor_workspace/common_workspace"
)

type Workspace struct {
	common_workspace.CommonWorkspace
}

func (w *Workspace) Initialize(host *engine.Host) {
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", nil, nil)
}

func (w *Workspace) Open()  { w.CommonOpen() }
func (w *Workspace) Close() { w.CommonClose() }
