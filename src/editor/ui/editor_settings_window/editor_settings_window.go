package editor_settings_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/editor/plugins"
	"kaiju/engine/host_container"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/engine/ui"
	"strconv"
)

type EditorSettingsWindow struct {
	uiMan            ui.Manager
	GoCompilerPath   string
	GridSnapping     float32
	RotationSnapping float32
}

func updateCompilerPath(elm *document.Element) {
	txt := elm.UI.ToInput().Text()
	editor_cache.SetEditorConfigValue(editor_cache.KaijuGoCompiler, txt)
}

func updateGridSnapping(elm *document.Element) {
	txt := elm.UI.ToInput().Text()
	if snap, err := strconv.ParseFloat(txt, 32); err == nil {
		editor_cache.SetEditorConfigValue(editor_cache.GridSnapping, snap)
	}
}

func updateRotationSnapping(elm *document.Element) {
	txt := elm.UI.ToInput().Text()
	if snap, err := strconv.ParseFloat(txt, 32); err == nil {
		editor_cache.SetEditorConfigValue(editor_cache.RotationSnapping, snap)
	}
}

func New() {
	const html = "editor/ui/editor_settings_window.html"
	esw := &EditorSettingsWindow{
		GridSnapping:     1,
		RotationSnapping: 15,
	}
	container := host_container.New("Editor Settings Window", nil)
	esw.uiMan.Init(container.Host)
	go container.Run(500, 300, -1, -1)
	<-container.PrepLock
	if v, ok := editor_cache.EditorConfigValue(editor_cache.KaijuGoCompiler); ok {
		esw.GoCompilerPath = v.(string)
	}
	if v, ok := editor_cache.EditorConfigValue(editor_cache.GridSnapping); ok {
		esw.GridSnapping = float32(v.(float64))
	}
	if v, ok := editor_cache.EditorConfigValue(editor_cache.RotationSnapping); ok {
		esw.RotationSnapping = float32(v.(float64))
	}
	container.RunFunction(func() {
		markup.DocumentFromHTMLAsset(&esw.uiMan, html, esw, map[string]func(*document.Element){
			"updateCompilerPath":     updateCompilerPath,
			"updateGridSnapping":     updateGridSnapping,
			"updateRotationSnapping": updateRotationSnapping,
			"regeneratePluginAPI":    func(*document.Element) { plugins.RegenerateAPI() },
		})
	})
}
