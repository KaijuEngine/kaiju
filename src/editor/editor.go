package editor

import (
	"kaiju/editor/ui/menu"
	"kaiju/editor/ui/project_window"
	"kaiju/engine"
)

type Editor struct {
	Host          *engine.Host
	menu          *menu.Menu
	projectWindow *project_window.ProjectWindow
}

func New(host *engine.Host) *Editor {
	host.SetFrameRateLimit(60)
	return &Editor{
		Host: host,
	}
}

func (e *Editor) SetupUI() {
	e.Host.CreatingEditorEntities()
	//e.menu = menu.New(e.Host)
	e.projectWindow, _ = project_window.New()
	<-e.projectWindow.Done
	e.Host.DoneCreatingEditorEntities()
}
