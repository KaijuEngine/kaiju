package vfx_workspace

import (
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type VfxWorkspace struct {
	common_workspace.CommonWorkspace
}

func (w *VfxWorkspace) Initialize(host *engine.Host) {
	defer tracing.NewRegion("VfxWorkspace.Initialize").End()
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/vfx_workspace.go.html", nil, map[string]func(*document.Element){})
}

func (w *VfxWorkspace) Open() {
	defer tracing.NewRegion("UIWorkspace.Open").End()
	w.CommonOpen()
}

func (w *VfxWorkspace) Close() {
	defer tracing.NewRegion("UIWorkspace.Close").End()
	w.CommonClose()
}

func (w *VfxWorkspace) Hotkeys() []common_workspace.HotKey {
	return []common_workspace.HotKey{}
}
