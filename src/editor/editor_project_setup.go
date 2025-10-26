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
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
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
	"fmt"
	"kaiju/editor/editor_overlay/new_project"
	"kaiju/editor/project"
	"kaiju/klib"
	"kaiju/platform/profiler/tracing"
	"log/slog"
)

func (ed *Editor) setProjectName(name string) {
	ed.host.Window.SetTitle(fmt.Sprintf("%s - Kaiju Engine Editor", name))
	ed.project.SetName(name)
}

func (ed *Editor) newProjectOverlay() {
	defer tracing.NewRegion("Editor.newProjectOverlay").End()
	new_project.Show(ed.host, new_project.Config{
		OnCreate: ed.createProject,
		OnOpen:   ed.openProject,
	})
}

func (ed *Editor) retryNewProjectOverlay(err error) {
	new_project.Show(ed.host, new_project.Config{
		OnCreate: ed.createProject,
		OnOpen:   ed.openProject,
		Error:    err.Error(),
	})
}

func (ed *Editor) createProject(name, path string) {
	defer tracing.NewRegion("Editor.createProject").End()
	err := ed.project.Initialize(path)
	if err != nil && !klib.ErrorIs[project.ConfigLoadError](err) {
		slog.Error("failed to create the project", "error", err)
		ed.retryNewProjectOverlay(err)
		return
	}
	ed.setProjectName(name)
	ed.lateLoadUI()
	ed.FocusInterface()
}

func (ed *Editor) openProject(path string) {
	defer tracing.NewRegion("Editor.openProject").End()
	if err := ed.project.Open(path); err != nil {
		slog.Error("failed to open the project", "error", err)
		ed.retryNewProjectOverlay(err)
		return
	}
	ed.setProjectName(ed.project.Name())
	ed.lateLoadUI()
	ed.FocusInterface()
}
