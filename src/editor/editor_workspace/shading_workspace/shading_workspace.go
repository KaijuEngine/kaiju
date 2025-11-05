package shading_workspace

import (
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type ShadingWorkspace struct {
	common_workspace.CommonWorkspace
}

func (w *ShadingWorkspace) Initialize(host *engine.Host) {
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/shading_workspace.go.html", nil, map[string]func(*document.Element){})
}

func (w *ShadingWorkspace) Open() {
	defer tracing.NewRegion("ShadingWorkspace.Open").End()
	w.CommonOpen()
}

func (w *ShadingWorkspace) Close() {
	defer tracing.NewRegion("ShadingWorkspace.Close").End()
	w.CommonClose()
}
