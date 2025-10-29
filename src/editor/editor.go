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
	"kaiju/editor/editor_embedded_content"
	"kaiju/editor/editor_events"
	"kaiju/editor/editor_logging"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_workspace"
	"kaiju/editor/editor_workspace/content_workspace"
	"kaiju/editor/editor_workspace/stage_workspace"
	"kaiju/editor/global_interface/menu_bar"
	"kaiju/editor/global_interface/status_bar"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/engine"
	"kaiju/engine/systems/events"
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
	window           struct {
		activateId     events.Id
		deactivateId   events.Id
		lastActiveTime time.Time
	}
}

type workspaces struct {
	stage   stage_workspace.Workspace
	content content_workspace.Workspace
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
}

func (ed *Editor) earlyLoadUI() {
	defer tracing.NewRegion("Editor.earlyLoadUI").End()
	ed.globalInterfaces.menuBar.Initialize(ed.host, ed)
	ed.globalInterfaces.statusBar.Initialize(ed.host, &ed.logging, ed)
}

func (ed *Editor) lateLoadUI() {
	defer tracing.NewRegion("Editor.lateLoadUI").End()
	slog.Info("compiling the project to get things ready")
	// Loose goroutine
	go ed.project.Compile()
	// Loose goroutine
	go ed.project.ReadSourceCode()
	ed.host.AssetDatabase().(*editor_embedded_content.EditorContent).Pfs = ed.project.FileSystem()
	ed.setupWindowActivity()
	ed.workspaces.stage.Initialize(ed.host, ed)
	ed.workspaces.content.Initialize(ed.host, &ed.events,
		ed.project.FileSystem(), ed.project.CacheDatabase())
	ed.setWorkspaceState(WorkspaceStateStage)
}
