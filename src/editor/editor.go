/******************************************************************************/
/* editor.go                                                                  */
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
	"kaiju/build"
	"kaiju/editor/editor_embedded_content"
	"kaiju/editor/editor_events"
	"kaiju/editor/editor_logging"
	"kaiju/editor/editor_overlay/ai_prompt"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_stage_manager/editor_stage_view"
	"kaiju/editor/editor_workspace"
	"kaiju/editor/editor_workspace/content_workspace"
	"kaiju/editor/editor_workspace/settings_workspace"
	"kaiju/editor/editor_workspace/shading_workspace"
	"kaiju/editor/editor_workspace/stage_workspace"
	"kaiju/editor/editor_workspace/ui_workspace"
	"kaiju/editor/global_interface/menu_bar"
	"kaiju/editor/global_interface/status_bar"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/engine"
	"kaiju/engine/systems/events"
	"kaiju/engine/ui"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"time"
)

// Editor is the entry point structure for the entire editor. It acts as the
// delegate to the various systems and holds the primary members that make up
// the bulk of the editor identity.
//
// The design goal of the editor is different than that of the [engine.Host], as
// it is not intended to be passed around for access to the system. Instead it
// will supply interface functions that are needed to the systems that it holds
// internally.
type Editor struct {
	host             *engine.Host
	settings         editor_settings.Settings
	project          project.Project
	workspaceState   WorkspaceState
	workspaces       workspaces
	globalInterfaces globalInterface
	currentWorkspace editor_workspace.Workspace
	logging          editor_logging.Logging
	history          memento.History
	events           editor_events.EditorEvents
	stageView        editor_stage_view.StageView
	window           struct {
		activateId     events.Id
		deactivateId   events.Id
		lastActiveTime time.Time
	}
	updateId engine.UpdateId
	blurred  bool
}

type workspaces struct {
	stage    stage_workspace.StageWorkspace
	content  content_workspace.ContentWorkspace
	shading  shading_workspace.ShadingWorkspace
	ui       ui_workspace.UIWorkspace
	settings settings_workspace.SettingsWorkspace
}

type globalInterface struct {
	menuBar   menu_bar.MenuBar
	statusBar status_bar.StatusBar
}

// FocusInterface is responsible for enabling the input on the various
// interfaces that are currently presented to the developer. This primarily
// includes the menu bar, status bar, and whichever workspace is active.
func (ed *Editor) FocusInterface() {
	defer tracing.NewRegion("Editor.FocusInterface").End()
	ed.globalInterfaces.menuBar.Focus()
	ed.globalInterfaces.statusBar.Focus()
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Focus()
	}
	ed.blurred = false
}

// FocusInterface is responsible for disabling the input on the various
// interfaces that are currently presented to the developer. This primarily
// includes the menu bar, status bar, and whichever workspace is active.
func (ed *Editor) BlurInterface() {
	defer tracing.NewRegion("Editor.BlurInterface").End()
	ed.globalInterfaces.menuBar.Blur()
	ed.globalInterfaces.statusBar.Blur()
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Blur()
	}
	ed.blurred = true
}

func (ed *Editor) earlyLoadUI() {
	defer tracing.NewRegion("Editor.earlyLoadUI").End()
	ed.globalInterfaces.menuBar.Initialize(ed.host, ed)
	ed.globalInterfaces.statusBar.Initialize(ed.host, &ed.logging, ed)
}

func (ed *Editor) UpdateSettings() {
	ed.host.SetFrameRateLimit(int64(klib.Clamp(ed.settings.RefreshRate, 0, 320)))
	if matrix.Approx(ed.settings.UIScrollSpeed, 0) {
		ed.settings.UIScrollSpeed = 1
	}
	ui.UIScrollSpeed = ed.settings.UIScrollSpeed
	if err := ed.settings.Save(); err != nil {
		slog.Error("failed to save the editor settings", "error", err)
		return
	}
}

func (ed *Editor) postProjectLoad() {
	defer tracing.NewRegion("Editor.lateLoadUI").End()
	ed.settings.AddRecentProject(ed.project.FileSystem().FullPath(""))
	slog.Info("compiling the project to get things ready")
	ed.host.AssetDatabase().(*editor_embedded_content.EditorContent).Pfs = ed.project.FileSystem()
	ed.setupWindowActivity()
	ed.workspaces.stage.Initialize(ed.host, ed)
	ed.workspaces.content.Initialize(ed.host, ed)
	ed.workspaces.shading.Initialize(ed.host, ed)
	ed.workspaces.ui.Initialize(ed.host, ed)
	ed.workspaces.settings.Initialize(ed.host, ed)
	ed.setWorkspaceState(WorkspaceStateStage)
	// goroutine
	go ed.project.CompileDebug()
	// goroutine
	go ed.project.ReadSourceCode()
	if build.Debug && ed.initAutoTest() {
		ed.updateId = ed.host.Updater.AddUpdate(ed.runAutoTest)
	} else {
		ed.updateId = ed.host.Updater.AddUpdate(ed.update)
	}
}

func (ed *Editor) update(deltaTime float64) {
	if ed.blurred {
		return
	}
	kb := &ed.host.Window.Keyboard
	if kb.HasCtrl() {
		if kb.KeyDown(hid.KeyboardKeyZ) {
			if !kb.HasShift() {
				ed.history.Undo()
			} else {
				ed.history.Redo()
			}
		} else if kb.KeyDown(hid.KeyboardKeyY) {
			ed.history.Redo()
		}
	}
	if kb.HasShift() && kb.KeyDown(hid.KeyboardKeyF1) {
		ed.blurred = true
		ed.BlurInterface()
		ai_prompt.Show(ed.host, func() {
			ed.FocusInterface()
		})
	}
	processWorkspaceHotkeys(ed, kb)
}

func processWorkspaceHotkeys(ed *Editor, kb *hid.Keyboard) {
	for _, hk := range ed.currentWorkspace.Hotkeys() {
		if hk.Ctrl && !kb.HasCtrl() {
			continue
		}
		if hk.Shift && !kb.HasShift() {
			continue
		}
		if hk.Alt && !kb.HasAlt() {
			continue
		}
		down := false
		valid := true
		for i := 0; i < len(hk.Keys) && valid; i++ {
			valid = kb.KeyHeld(hk.Keys[i])
			down = down || kb.KeyDown(hk.Keys[i])
		}
		if valid && down {
			hk.Call()
		}
	}
}
