/******************************************************************************/
/* new_project_overlay.go                                                     */
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

package new_project

import (
	"kaiju/editor/editor_overlay/file_browser"
	"kaiju/engine"
	"kaiju/engine/ui"
	"kaiju/engine/ui/markup"
	"kaiju/engine/ui/markup/document"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"path/filepath"
)

type NewProject struct {
	doc       *document.Document
	uiMan     ui.Manager
	nameInput *document.Element
	folder    *document.Element
	config    Config
}

type Config struct {
	// OnCreate will be called when the "Create" button is clicked, it will
	// return the name that the developer typed in and the path they selected.
	OnCreate func(name, path string)

	// OnOpen will be called when the "Browse" button is clicked, it will return
	// the path that was selected.
	OnOpen func(string)

	// Error will be used to print out an error to the developer in the window.
	Error string

	// RecentProjects is a list of paths to recent projects.
	RecentProjects []string
}

type OverlayData struct {
	Error          string
	RecentProjects []struct {
		Path  string
		Label string
	}
}

func Show(host *engine.Host, config Config) (*NewProject, error) {
	defer tracing.NewRegion("new_project.Show").End()
	np := &NewProject{config: config}
	np.uiMan.Init(host)
	var err error
	data := OverlayData{Error: config.Error}
	for i := range config.RecentProjects {
		data.RecentProjects = append(data.RecentProjects, struct {
			Path  string
			Label string
		}{
			Path:  config.RecentProjects[i],
			Label: filepath.Base(config.RecentProjects[i]),
		})
	}
	np.doc, err = markup.DocumentFromHTMLAsset(&np.uiMan,
		"editor/ui/overlay/new_project_overlay.go.html",
		data, map[string]func(*document.Element){
			"openProject":       np.openProject,
			"browse":            np.browse,
			"createProject":     np.createProject,
			"openRecentProject": np.openRecentProject,
		})
	if err != nil {
		return np, err
	}
	np.nameInput, _ = np.doc.GetElementById("nameInput")
	np.folder, _ = np.doc.GetElementById("folder")
	return np, err
}

func (np *NewProject) Close() {
	defer tracing.NewRegion("NewProject.Close").End()
	np.doc.Destroy()
}

func (np *NewProject) openProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.openProject").End()
	np.showFolderPick(true)
}

func (np *NewProject) browse(e *document.Element) {
	defer tracing.NewRegion("NewProject.createFolder").End()
	np.showFolderPick(false)
}

func (np *NewProject) showFolderPick(isOpen bool) {
	defer tracing.NewRegion("NewProject.showFolderPick").End()
	np.uiMan.DisableUpdate()
	file_browser.Show(np.uiMan.Host, file_browser.Config{
		OnlyFolders: true,
		OnConfirm: func(paths []string) {
			np.uiMan.EnableUpdate()
			if isOpen {
				np.openProjectFolder(paths[0])
			} else {
				np.folder.UI.ToInput().SetText(paths[0])
			}
		}, OnCancel: func() {
			np.uiMan.EnableUpdate()
		},
	})
}

func (np *NewProject) createProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.createProject").End()
	name := np.nameInput.UI.ToInput().Text()
	path := np.folder.UI.ToInput().Text()
	if name == "" {
		slog.Error("project name was not set")
		return
	}
	if path == "" {
		slog.Error("project path was not set")
	}
	np.Close()
	if np.config.OnCreate == nil {
		slog.Error("nothing bound to OnCreate, doing nothing")
		return
	}
	np.config.OnCreate(name, path)
}

func (np *NewProject) openRecentProject(e *document.Element) {
	defer tracing.NewRegion("NewProject.openRecentProject").End()
	np.uiMan.Host.RunOnMainThread(func() {
		np.openProjectFolder(e.Attribute("data-path"))
	})
}

func (np *NewProject) openProjectFolder(path string) {
	np.Close()
	if np.config.OnOpen == nil {
		slog.Error("nothing bound to OnOpen, doing nothing")
		return
	}
	np.config.OnOpen(path)
}
