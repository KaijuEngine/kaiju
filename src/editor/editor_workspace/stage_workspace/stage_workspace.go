/******************************************************************************/
/* stage_workspace.go                                                         */
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

package stage_workspace

import (
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_workspace/common_workspace"
	"kaiju/editor/editor_workspace/content_workspace"
	"kaiju/editor/editor_workspace/stage_workspace/transform_tools"
	"kaiju/editor/memento"
	"kaiju/editor/project/project_database/content_database"
	"kaiju/editor/project/project_file_system"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/ui/markup/document"
	"kaiju/matrix"
	"kaiju/rendering"
	"log/slog"
)

const maxContentDropDistance = 10

type Workspace struct {
	common_workspace.CommonWorkspace
	camera        editor_controls.EditorCamera
	updateId      engine.UpdateId
	gridShader    *rendering.ShaderDataBasic
	pageData      content_workspace.WorkspaceUIData
	pfs           *project_file_system.FileSystem
	cdb           *content_database.Cache
	contentUI     WorkspaceContentUI
	manager       editor_stage_manager.StageManager
	transformTool transform_tools.TransformTool
}

func (w *Workspace) WorkspaceHost() *engine.Host { return w.Host }

func (w *Workspace) Manager() *editor_stage_manager.StageManager { return &w.manager }

func (w *Workspace) Camera() *editor_controls.EditorCamera { return &w.camera }

func (w *Workspace) Initialize(host *engine.Host, history *memento.History, pfs *project_file_system.FileSystem, cdb *content_database.Cache) {
	w.pfs = pfs
	w.cdb = cdb
	w.manager.Initialize(host)
	ids := w.pageData.SetupUIData(cdb)
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/stage_workspace.go.html", w.pageData, map[string]func(*document.Element){
			"inputFilter":    w.contentUI.inputFilter,
			"tagFilter":      w.contentUI.tagFilter,
			"clickFilter":    w.contentUI.clickFilter,
			"hideContent":    w.contentUI.hideContent,
			"showContent":    w.contentUI.showContent,
			"entryDragStart": w.contentUI.entryDragStart,
		})
	w.createViewportGrid()
	w.setupCamera()
	w.contentUI.setup(w, ids)
	w.transformTool.Initialize(host, w, history)
}

func (w *Workspace) Open() {
	w.CommonOpen()
	w.gridShader.Activate()
	w.updateId = w.Host.Updater.AddUpdate(w.update)
	w.contentUI.open()
	w.Host.RunOnMainThread(w.Doc.Clean)
}

func (w *Workspace) Close() {
	w.Host.Updater.RemoveUpdate(&w.updateId)
	w.gridShader.Deactivate()
	w.CommonClose()
}

func (w *Workspace) update(deltaTime float64) {
	if !w.contentUI.update(w) {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.contentUI.processHotkeys(w.Host)
	if !w.camera.Update(w.Host, deltaTime) {
		w.processViewportInteractions()
	}
}

func (w *Workspace) createViewportGrid() {
	const gridCount = 100
	const halfGridCount = gridCount / 2
	material, err := w.Host.MaterialCache().Material(assets.MaterialDefinitionGrid)
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
	grid := rendering.NewMeshGrid(w.Host.MeshCache(), "viewport_grid",
		points, matrix.Color{0.5, 0.5, 0.5, 1})
	w.gridShader = &rendering.ShaderDataBasic{
		ShaderDataBase: rendering.NewShaderDataBase(),
		Color:          matrix.Color{0.5, 0.5, 0.5, 1},
	}
	w.Host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   w.Host.Window.Renderer,
		Material:   material,
		Mesh:       grid,
		ShaderData: w.gridShader,
	})
}

func (w *Workspace) setupCamera() {
	w.camera.OnModeChange.Add(func() {
		m := matrix.Mat4Identity()
		switch w.camera.Mode() {
		case editor_controls.EditorCameraMode3d:
			// Identity matrix is fine
		case editor_controls.EditorCameraMode2d:
			m.RotateX(90)
		}
		w.gridShader.SetModel(m)
	})
	w.camera.SetMode(editor_controls.EditorCameraMode3d, w.Host)
}
