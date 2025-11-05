package settings_workspace

import (
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
)

type SettingsWorkspace struct {
	common_workspace.CommonWorkspace
}

func (w *SettingsWorkspace) Initialize(host *engine.Host) {
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/settings_workspace.go.html", nil, map[string]func(*document.Element){})
}

func (w *SettingsWorkspace) Open() {
	defer tracing.NewRegion("SettingsWorkspace.Open").End()
	w.CommonOpen()
}

func (w *SettingsWorkspace) Close() {
	defer tracing.NewRegion("SettingsWorkspace.Close").End()
	w.CommonClose()
}
