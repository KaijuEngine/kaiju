package content_workspace

import (
	"kaiju/engine"
	"kaiju/engine/ui/markup/document"
	"kaiju/games/editor/editor_overlay/file_browser"
	"kaiju/games/editor/editor_workspace/common_workspace"
	"kaiju/games/editor/project/project_database/content_database"
	"kaiju/games/editor/project/project_file_system"
	"log/slog"
)

type Workspace struct {
	common_workspace.CommonWorkspace
	pfs *project_file_system.FileSystem
}

func (w *Workspace) Initialize(host *engine.Host, pfs *project_file_system.FileSystem) {
	w.pfs = pfs
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/content_workspace.go.html", nil, map[string]func(*document.Element){
			"clickImport": w.clickImport,
		})
}

func (w *Workspace) Open()  { w.CommonOpen() }
func (w *Workspace) Close() { w.CommonClose() }

func (w *Workspace) clickImport(*document.Element) {
	w.UiMan.DisableUpdate()
	file_browser.Show(w.Host, file_browser.Config{
		ExtFilter:   content_database.ImportableTypes,
		MultiSelect: true,
		OnConfirm: func(paths []string) {
			w.UiMan.EnableUpdate()
			for i := range paths {
				_, err := content_database.Import(paths[i], w.pfs)
				if err != nil {
					slog.Error("failed to import content", "path", paths[i], "error", err)
				}
			}
		}, OnCancel: func() {
			w.UiMan.EnableUpdate()
		},
	})
}
