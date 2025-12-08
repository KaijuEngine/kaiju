/******************************************************************************/
/* editor_menu_bar_handler.go                                                 */
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
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_overlay/input_prompt"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// StageWorkspaceSelected will inform the editor that the developer has
// changed to the stage workspace. This is an exposed function to meet the
// interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) StageWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateStage)
}

// ContentWorkspaceSelected will inform the editor that the developer has
// changed to the content workspace. This is an exposed function to meet the
// interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) ContentWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateContent)
}

// ShadingWorkspaceSelected will inform the editor that the developer has
// changed to the shading workspace. This is an exposed function to meet the
// interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) ShadingWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateShading)
}

// UIWorkspaceSelected will inform the editor that the developer has changed to
// the ui workspace. This is an exposed function to meet the interface needs of
// [menu_bar.MenuBarHandler].
func (ed *Editor) UIWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateUI)
}

// SettingsWorkspaceSelected will inform the editor that the developer has
// changed to the settings workspace. This is an exposed function to meet the
// interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) SettingsWorkspaceSelected() {
	ed.setWorkspaceState(WorkspaceStateSettings)
}

func (ed *Editor) Build() {
	if !ed.ensureMainStageExists() {
		return
	}
	// goroutine
	go ed.project.CompileDebug()
	// goroutine
	go ed.project.Package()
}

func (ed *Editor) BuildAndRun() {
	if !ed.ensureMainStageExists() {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	// goroutine
	go func() {
		ed.project.CompileDebug()
		wg.Done()
	}()
	// goroutine
	go func() {
		ed.project.Package()
		wg.Done()
	}()
	// goroutine
	go func() {
		wg.Wait()
		ed.project.Run()
	}()
}

func (ed *Editor) BuildAndRunCurrentStage() {
	stageId := ed.stageView.Manager().StageId()
	if stageId == "" {
		slog.Error("current stage has not yet been created, please save it to test")
		return
	}
	ed.stageView.Manager().SaveStage(ed.Cache(), ed.project.FileSystem())
	wg := sync.WaitGroup{}
	wg.Add(2)
	// goroutine
	go func() {
		ed.project.CompileDebug()
		wg.Done()
	}()
	// goroutine
	go func() {
		ed.project.Package()
		wg.Done()
	}()
	// goroutine
	go func() {
		wg.Wait()
		ed.project.Run("-startStage", stageId)
	}()
}

// OpenVSCodeProject will open Visual Studio Code directly to the project top-
// level folder. This is an exposed function to meet the interface needs of
// [menu_bar.MenuBarHandler].
func (ed *Editor) OpenVSCodeProject() {
	defer tracing.NewRegion("Editor.OpenVSCodeProject").End()
	exec.Command("code", ed.project.FileSystem().FullPath("")).Run()
}

func (ed *Editor) CreateNewStage() {
	if ed.history.HasPendingChanges() {
		ed.BlurInterface()
		confirm_prompt.Show(ed.host, confirm_prompt.Config{
			Title:       "Discard changes",
			Description: "You have unsaved changes to your stage, would you like to discard them and create a new stage?",
			ConfirmText: "Yes",
			CancelText:  "No",
			OnConfirm: func() {
				ed.FocusInterface()
				ed.stageView.Manager().NewStage()
			},
			OnCancel: func() { ed.FocusInterface() },
		})
	} else {
		ed.stageView.Manager().NewStage()
	}
}

// SaveCurrentStage will save the currently open stage file. This is an exposed
// function to meet the interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) SaveCurrentStage() {
	defer tracing.NewRegion("Editor.SaveCurrentStage").End()
	sm := ed.stageView.Manager()
	if sm.IsNew() {
		ed.BlurInterface()
		input_prompt.Show(ed.host, input_prompt.Config{
			Title:       "Name stage",
			Description: "What would you like to name your stage?",
			Placeholder: "Stage name...",
			Value:       "New Stage",
			ConfirmText: "Save",
			CancelText:  "Cancel",
			OnConfirm: func(name string) {
				ed.FocusInterface()
				ed.saveNewStage(strings.TrimSpace(name))
			},
			OnCancel: func() { ed.FocusInterface() },
		})
	} else {
		ed.saveCurrentStageWithoutNameInput()
	}
}

func (ed *Editor) CreateNewCamera() {
	ed.workspaces.stage.CreateNewCamera()
}

func (ed *Editor) CreateNewEntity() {
	ed.history.BeginTransaction()
	e, _ := ed.workspaces.stage.CreateNewEntity()
	m := ed.stageView.Manager()
	m.ClearSelection()
	m.SelectEntity(e)
	ed.history.CommitTransaction()
}

func (ed *Editor) CreateNewLight() {
	ed.workspaces.stage.CreateNewLight()
}

func (ed *Editor) CreateHtmlUiFile(name string) {
	tplPath := filepath.Join(project_file_system.ProjectFileTemplates,
		"html_file_template.html.txt")
	data, err := ed.project.FileSystem().ReadFile(tplPath)
	if err != nil {
		slog.Error("failed to read the html template file", "file", tplPath, "error", err)
		return
	}
	t, err := template.New("Html").Parse(string(data))
	if err != nil {
		slog.Error("failed to parse the html template file", "file", tplPath, "error", err)
		return
	}
	sb := strings.Builder{}
	if err = t.Execute(&sb, name); err != nil {
		slog.Error("failed to execute the html template", "file", tplPath, "error", err)
		return
	}
	pfs := ed.ProjectFileSystem()
	cache := ed.Cache()
	ids := content_database.ImportRaw(name, []byte(sb.String()), content_database.Html{}, pfs, cache)
	if len(ids) > 0 {
		ed.events.OnContentAdded.Execute(ids)
		cc, err := cache.Read(ids[0])
		if err != nil {
			exec.Command("code", pfs.FullPath(""), pfs.FullPath(cc.ContentPath())).Run()
		}
	}
}

func (ed *Editor) saveCurrentStageWithoutNameInput() {
	sm := ed.stageView.Manager()
	if err := sm.SaveStage(ed.project.CacheDatabase(), ed.project.FileSystem()); err == nil {
		ed.history.SetSavePosition()
	} else {
		slog.Error("failed to save the current stage", "error", err)
	}
}

func (ed *Editor) ensureMainStageExists() bool {
	if ed.project.Settings().EntryPointStage == "" {
		slog.Error("failed to build, 'main stage' not set in project settings")
		return false
	}
	return true
}

func (ed *Editor) saveNewStage(name string) {
	if name == "" {
		slog.Error("name was blank for the stage, can't save")
		return
	}
	sm := ed.stageView.Manager()
	if err := sm.SetStageId(name, ed.Cache()); err != nil {
		slog.Error("failed to save stage", "error", err)
		return
	}
	ed.saveCurrentStageWithoutNameInput()
	id := sm.StageId()
	ed.events.OnContentAdded.Execute([]string{id})
	// If the entry point stage hasn't yet been created in the
	// settings, assume that this stage will be the one.
	ps := ed.project.Settings()
	if ps.EntryPointStage == "" {
		ps.EntryPointStage = id
		ps.Save(ed.project.FileSystem())
		ed.workspaces.settings.RequestReload()
	}
}
