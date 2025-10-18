package editor

import (
	"fmt"
	"kaiju/games/editor/editor_overlay/new_project"
	"kaiju/games/editor/project"
	"kaiju/platform/profiler/tracing"
	"log/slog"
)

func (ed *Editor) newProjectOverlay() {
	defer tracing.NewRegion("Editor.newProjectOverlay").End()
	new_project.Show(ed.host, new_project.Config{
		OnCreate: ed.createProject,
		OnOpen:   ed.openProject,
	})
}

func (ed *Editor) createProject(name, path string) {
	defer tracing.NewRegion("Editor.createProject").End()
	err := ed.project.Initialize(path)
	if _, ok := err.(project.ConfigLoadError); !ok {
		// TODO:  The new project window is gone by this point, it neesd to
		// come back in the case of a failure here
		slog.Error("failed to create the project", "error", err)
		return
	}
	ed.SetProjectName(name)
	ed.lateLoadUI()
	ed.focusInterface()
}

func (ed *Editor) SetProjectName(name string) {
	ed.host.Window.SetTitle(fmt.Sprintf("%s - Kaiju Engine Editor", name))
	ed.project.SetName(name)
}

func (ed *Editor) openProject(path string) {
	defer tracing.NewRegion("Editor.openProject").End()
	if err := ed.project.Open(path); err != nil {
		slog.Error("failed to create the project", "error", err)
		return
	}
	ed.SetProjectName(ed.project.Name())
	ed.lateLoadUI()
	ed.focusInterface()
}
