package ui_workspace

import (
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type UIWorkspace struct {
	common_workspace.CommonWorkspace
}

func (w *UIWorkspace) Initialize(host *engine.Host) {
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/ui_workspace.go.html", nil, map[string]func(*document.Element){})
}

func (w *UIWorkspace) Open() {
	defer tracing.NewRegion("UIWorkspace.Open").End()
	w.CommonOpen()
}

func (w *UIWorkspace) Close() {
	defer tracing.NewRegion("UIWorkspace.Close").End()
	w.CommonClose()
}
