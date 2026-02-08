/******************************************************************************/
/* editor_project_setup.go                                                    */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor

import (
	"errors"
	"fmt"
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_overlay/new_project"
	"kaiju/editor/project"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
)

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
	forceReplace := kb.HasShift() || kb.HasCtrl()
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
