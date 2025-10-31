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
	"kaiju/editor/editor_overlay/confirm_prompt"
	"kaiju/editor/editor_overlay/input_prompt"
	"kaiju/editor/editor_stage_manager"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
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

func (ed *Editor) Build() {
	if !ed.ensureMainStageExists() {
		return
	}
	// Loose goroutine
	go ed.project.CompileDebug()
	// Loose goroutine
	go ed.project.Package()
}

func (ed *Editor) BuildAndRun() {
	if !ed.ensureMainStageExists() {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	// Loose goroutine
	go func() {
		ed.project.CompileDebug()
		wg.Done()
	}()
	// Loose goroutine
	go func() {
		ed.project.Package()
		wg.Done()
	}()
	// Loose goroutine
	go func() {
		wg.Wait()
		ed.project.Run()
	}()
}

func (ed *Editor) BuildAndRunCurrentStage() {
	stageId := ed.workspaces.stage.Manager().StageId()
	if stageId == "" {
		slog.Error("current stage has not yet been created, please save it to test")
		return
	}
	ed.workspaces.stage.Manager().SaveStage(ed.Cache(), ed.project.FileSystem())
	wg := sync.WaitGroup{}
	wg.Add(2)
	// Loose goroutine
	go func() {
		ed.project.CompileDebug()
		wg.Done()
	}()
	// Loose goroutine
	go func() {
		ed.project.Package()
		wg.Done()
	}()
	// Loose goroutine
	go func() {
		wg.Wait()
		ed.project.Run(fmt.Sprintf("-startStage=%s", stageId))
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
			Title:       "Discrad changes",
			Description: "You have unsaved changes to your stage, would you like to discard them and create a new stage?",
			ConfirmText: "Yes",
			CancelText:  "No",
			OnConfirm: func() {
				ed.FocusInterface()
				ed.workspaces.stage.Manager().NewStage()
			},
			OnCancel: func() { ed.FocusInterface() },
		})
	} else {
		ed.workspaces.stage.Manager().NewStage()
	}
}

// SaveCurrentStage will save the currently open stage file. This is an exposed
// function to meet the interface needs of [menu_bar.MenuBarHandler].
func (ed *Editor) SaveCurrentStage() {
	defer tracing.NewRegion("Editor.SaveCurrentStage").End()
	sm := ed.workspaces.stage.Manager()
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
				name = strings.TrimSpace(name)
				if name != "" {
					if err := sm.SetStageId(name, ed.Cache()); err != nil {
						slog.Error("failed to save stage", "error", err)
						return
					}
					ed.saveCurrentStageWithoutNameInput()
					id := sm.StageId()
					ed.workspaces.content.AddContent([]string{id})
				} else {
					slog.Error("name was blank for the stage, can't save")
				}
			},
			OnCancel: func() { ed.FocusInterface() },
		})
	} else {
		ed.saveCurrentStageWithoutNameInput()
	}
}

func (ed *Editor) saveCurrentStageWithoutNameInput() {
	sm := ed.workspaces.stage.Manager()
	if err := sm.SaveStage(ed.project.CacheDatabase(), ed.project.FileSystem()); err == nil {
		ed.history.SetSavePosition()
	} else {
		slog.Error("failed to save the current stage", "error", err)
	}
}

func (ed *Editor) ensureMainStageExists() bool {
	if _, err := ed.Cache().Read(editor_stage_manager.StageIdPrefix + "main"); err != nil {
		slog.Error("failed to build, no stage named 'main' found")
		return false
	}
	return true
}
