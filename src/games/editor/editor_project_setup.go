package editor

import (
	"kaiju/games/editor/editor_overlay/new_project"
	"log/slog"
)

func (ed *Editor) newProjectOverlay() {
	new_project.Show(ed.host, new_project.Config{
		OnCreate: ed.createProject,
		OnOpen:   ed.openProject,
	})
}

func (ed *Editor) createProject(name, path string) {
	if err := ed.setProject(path); err != nil {
		return
	}
	ed.project.SetName(name)
	ed.loadInterface()
}

func (ed *Editor) openProject(path string) {
	ed.setProject(path)
	ed.loadInterface()
}

func (ed *Editor) setProject(path string) error {
	err := ed.project.Initialize(path)
	if err != nil {
		slog.Error("failed to initialize the project", "error", err)
		return err
	}
	return nil
}
