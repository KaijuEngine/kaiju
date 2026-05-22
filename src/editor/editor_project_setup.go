/******************************************************************************/
/* editor_project_setup.go                                                    */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor

import (
	"errors"
	"fmt"
	"log/slog"

	"kaijuengine.com/build"
	"kaijuengine.com/editor/editor_overlay/confirm_prompt"
	"kaijuengine.com/editor/editor_overlay/new_project"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/profiler/tracing"
)

func CreateNewProjectFromCLI(path string) {
	if build.Editor {
		proj := project.Project{}
		templatePath := engine.LaunchParams.ProjectTemplate
		if err := proj.Initialize(path, templatePath, EditorVersion); err != nil {
			slog.Error("failed to create the project", "error", err, "path", path)
			return
		}
		if name := engine.LaunchParams.ProjectName; name != "" {
			proj.SetName(name)
		}
		if err := proj.Close(); err != nil {
			slog.Error("failed to save the project configuration", "error", err)
			return
		}
		slog.Info("successfully created blank project", "path", path)
	} else {
		slog.Error("the -newproject flag is only available in editor builds")
	}
}

func (ed *Editor) setProjectName(name string) {
	ed.host.RunOnMainThread(func() {
		ed.host.Window.SetTitle(fmt.Sprintf("%s - Kaiju Engine Editor", name))
	})
	ed.project.SetName(name)
}

func (ed *Editor) newProjectOverlay() {
	defer tracing.NewRegion("Editor.newProjectOverlay").End()
	new_project.Show(ed.host, new_project.Config{
		OnCreate:       ed.createProject,
		OnOpen:         ed.openProject,
		RecentProjects: ed.settings.RecentProjects,
	})
}

func (ed *Editor) retryNewProjectOverlay(err error) {
	new_project.Show(ed.host, new_project.Config{
		OnCreate:       ed.createProject,
		OnOpen:         ed.openProject,
		RecentProjects: ed.settings.RecentProjects,
		Error:          err.Error(),
	})
}

func (ed *Editor) createProject(name, path, templatePath string) {
	defer tracing.NewRegion("Editor.createProject").End()
	err := ed.project.Initialize(path, templatePath, EditorVersion)
	if err != nil && !klib.ErrorIs[project.ConfigLoadError](err) {
		slog.Error("failed to create the project", "error", err)
		ed.retryNewProjectOverlay(err)
		return
	}
	ed.setProjectName(name)
	ed.postProjectLoad()
	ed.FocusInterface()
}

func (ed *Editor) openProject(path string) {
	defer tracing.NewRegion("Editor.openProject").End()
	if err := ed.project.Open(path); err != nil {
		slog.Error("failed to open the project", "error", err)
		lastCount := len(ed.settings.RecentProjects)
		ed.settings.RecentProjects = klib.SlicesRemoveElement(ed.settings.RecentProjects, path)
		if len(ed.settings.RecentProjects) != lastCount {
			ed.settings.Save()
		}
		ed.retryNewProjectOverlay(err)
		return
	}
	projectVersion := ed.project.Settings.EditorVersion
	finishLoad := func() {
		ed.setProjectName(ed.project.Name())
		ed.postProjectLoad()
		ed.FocusInterface()
	}
	hasEngineSource := ed.project.FileSystem().HasEngineCode()
	// This is a special hidden feature for editor/engine developers to be able
	// to force updating engine code in projects. This makes it easier than
	// bumping the engine version to do the same thing (or deleting kaiju src)
	kb := &ed.host.Window.Keyboard
	forceReplace := kb.HasShift() || kb.HasCtrlOrMeta()
	if projectVersion != EditorVersion || !hasEngineSource || forceReplace {
		title := "Upgrade project"
		description := "Your project is for an older version of the editor, would you like to upgrade it? Please make sure you've backed up your project (with VCS for example) before proceeding."
		cancelMsg := "Project upgrade refused, unable to open project"
		if projectVersion == EditorVersion {
			title = "Import engine code"
			description = "Your project doesn't have the engine source, would you like to import it? This is typical if you don't commit the `kaiju` folder to your repository."
			cancelMsg = "Engine source import refused, unable to open project"
		}
		confirm_prompt.Show(ed.host, confirm_prompt.Config{
			Title:       title,
			Description: description,
			ConfirmText: "Yes",
			CancelText:  "Cancel",
			OnConfirm: func() {
				if err := ed.project.TryUpgrade(); err != nil {
					ed.retryNewProjectOverlay(err)
				} else {
					ed.project.Settings.EditorVersion = EditorVersion
					ed.project.Settings.Save(ed.ProjectFileSystem())
					finishLoad()
				}
			},
			OnCancel: func() {
				ed.retryNewProjectOverlay(errors.New(cancelMsg))
			},
		})
	} else {
		finishLoad()
	}
}
