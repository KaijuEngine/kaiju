/******************************************************************************/
/* editor_initialization.go                                                   */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
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
	"errors"
	"kaiju/editor/ui/content_details_window"
	"kaiju/editor/ui/content_window"
	"kaiju/editor/ui/context_menu"
	"kaiju/editor/ui/details_window"
	"kaiju/editor/ui/editor_menu"
	"kaiju/editor/ui/editor_window"
	"kaiju/editor/ui/hierarchy"
	"kaiju/editor/ui/log_window"
	"kaiju/editor/ui/project_window"
	"kaiju/editor/ui/status_bar"
	"kaiju/editor/ui/tab_container"
	"kaiju/editor/ui/viewport_overlay"
	"kaiju/editor/viewport/controls"
	"kaiju/editor/viewport/tools/transform_tools"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/host_container"
	"kaiju/engine/systems/logging"
	"kaiju/matrix"
	"kaiju/platform/profiler"
	"kaiju/rendering"
	tests "kaiju/tests/rendering_tests"
	"kaiju/tools/html_preview"
	"log/slog"
)

func addConsole(ed *Editor) {
	host := ed.container.Host
	html_preview.SetupConsole(host)
	profiler.SetupConsole(host)
	tests.SetupConsole(host)
	setupConsole(ed)
}

func setupEditorWindow(ed *Editor, logStream *logging.LogStream) {
	ed.container = host_container.New("Kaiju Editor", logStream)
	ed.container.Host.InitializeAudio()
	editor_window.OpenWindow(ed,
		engine.DefaultWindowWidth, engine.DefaultWindowHeight, -1, -1)
	ed.RunOnHost(func() { addConsole(ed) })
}

func waitForProjectSelectWindow(ed *Editor) (string, error) {
	cx, cy := ed.Host().Window.Center()
	projectWindow, _ := project_window.New(projectTemplate, cx, cy)
	projectPath := <-projectWindow.Selected
	if projectPath == "" {
		ed.Host().Close()
		return "", errors.New("invalid project path selected")
	}
	return projectPath, nil
}

func constructEditorUI(ed *Editor) {
	ed.Host().CreatingEditorEntities()
	ed.logWindow = log_window.New(ed.Host().LogStream, ed.ReloadTabs)
	ed.contentWindow = content_window.New(&ed.contentOpener, ed)
	ed.detailsWindow = details_window.New(ed)
	ed.contentDetailsWindow = content_details_window.New(ed)
	ed.contextMenu = context_menu.New(ed.container, &ed.uiManager)
	ed.hierarchy = hierarchy.New(ed.Host(), &ed.selection,
		hierarchyContextMenuActions(ed), ed.ReloadTabs)
	ed.menu = editor_menu.New(ed.container, ed.logWindow, ed.contentWindow,
		ed.hierarchy, &ed.contentOpener, ed, &ed.uiManager)
	ed.statusBar = status_bar.New(&ed.uiManager, func() {
		ed.ReloadOrOpenTab("Log")
	})
	viewport_overlay.New(ed, &ed.uiManager)
	setupViewportGrid(ed)
	{
		// TODO:  Testing tools
		ed.transformTool = transform_tools.New(ed.Host(), ed, "editor_overlay", &ed.history)
		ed.selection.Changed.Add(func() { ed.transformTool.Disable() })
	}
	ed.Host().DoneCreatingEditorEntities()
	ed.camera.SetMode(controls.EditorCameraMode3d, ed.Host())

	leftContainer := tab_container.New(ed.container.Host, &ed.uiManager,
		[]tab_container.TabContainerTab{
			tab_container.NewTab(ed.hierarchy),
		}, tab_container.SnapLeft)
	bottomContainer := tab_container.New(ed.container.Host, &ed.uiManager,
		[]tab_container.TabContainerTab{
			tab_container.NewTab(ed.contentWindow),
			tab_container.NewTab(ed.logWindow),
		}, tab_container.SnapBottom)
	rightContainer := tab_container.New(ed.container.Host, &ed.uiManager,
		[]tab_container.TabContainerTab{
			tab_container.NewTab(ed.detailsWindow),
			tab_container.NewTab(ed.contentDetailsWindow),
		}, tab_container.SnapRight)
	ed.tabContainers = append(ed.tabContainers, leftContainer)
	ed.tabContainers = append(ed.tabContainers, rightContainer)
	ed.tabContainers = append(ed.tabContainers, bottomContainer)
}

func setupViewportGrid(ed *Editor) {
	const gridCount = 20
	const halfGridCount = gridCount / 2
	host := ed.Host()
	material, err := host.MaterialCache().Material(assets.MaterialDefinitionGrid)
	if err != nil {
		slog.Error("failed to load the grid material", "error", err)
		return
	}
	points := make([]matrix.Vec3, 0, gridCount*4)
	for i := -halfGridCount; i <= halfGridCount; i++ {
		fi := float32(i)
		points = append(points, matrix.Vec3{fi, 0, -halfGridCount})
		points = append(points, matrix.Vec3{fi, 0, halfGridCount})
		points = append(points, matrix.Vec3{-halfGridCount, 0, fi})
		points = append(points, matrix.Vec3{halfGridCount, 0, fi})
	}
	grid := rendering.NewMeshGrid(host.MeshCache(), "viewport_grid",
		points, matrix.Color{0.5, 0.5, 0.5, 1})
	sd := &rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.Color{0.5, 0.5, 0.5, 1},
	}
	host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   host.Window.Renderer,
		Material:   material,
		Mesh:       grid,
		ShaderData: sd,
	})
	ed.camera.OnModeChange.Add(func() {
		m := matrix.Mat4Identity()
		switch ed.camera.Mode() {
		case controls.EditorCameraMode3d:
			// Identity matrix is fine
		case controls.EditorCameraMode2d:
			m.RotateX(90)
		}
		sd.SetModel(m)
	})
}
