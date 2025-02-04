package editor_settings_window

import (
	"kaiju/editor/cache/editor_cache"
	"kaiju/host_container"
	"kaiju/markup"
	"kaiju/markup/document"
	"kaiju/ui"
)

type EditorSettingsWindow struct {
	GoCompilerPath string
}

func updateCompilerPath(elm *document.Element) {
	txt := elm.UI.(*ui.Input).Text()
	editor_cache.SetEditorConfigValue(editor_cache.KaijuGoCompiler, txt)
}

func New() {
	const html = "editor/ui/editor_settings_window.html"
	container := host_container.New("Editor Settings Window", nil)
	go container.Run(500, 300, -1, -1)
	<-container.PrepLock
	esw := &EditorSettingsWindow{}
	if v, ok := editor_cache.EditorConfigValue(editor_cache.KaijuGoCompiler); ok {
		esw.GoCompilerPath = v.(string)
	}
	container.RunFunction(func() {
		markup.DocumentFromHTMLAsset(container.Host, html, esw, map[string]func(*document.Element){
			"updateCompilerPath": updateCompilerPath,
		})
	})
}
