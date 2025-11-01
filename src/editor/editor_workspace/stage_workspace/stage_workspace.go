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
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/engine/ui/markup/document"
	"kaiju/klib"
	"kaiju/matrix"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
)

const maxContentDropDistance = 10

type Workspace struct {
	common_workspace.CommonWorkspace
	ed            StageWorkspaceEditorInterface
	camera        editor_controls.EditorCamera
	updateId      engine.UpdateId
	gridTransform matrix.Transform
	gridShader    *shader_data_registry.ShaderDataGrid
	pageData      content_workspace.WorkspaceUIData
	contentUI     WorkspaceContentUI
	hierarchyUI   WorkspaceHierarchyUI
	detailsUI     WorkspaceDetailsUI
	manager       editor_stage_manager.StageManager
	transformTool transform_tools.TransformTool
}

func (w *Workspace) WorkspaceHost() *engine.Host { return w.Host }

func (w *Workspace) Manager() *editor_stage_manager.StageManager { return &w.manager }

func (w *Workspace) Camera() *editor_controls.EditorCamera { return &w.camera }

func (w *Workspace) Initialize(host *engine.Host, ed StageWorkspaceEditorInterface) {
	defer tracing.NewRegion("StageWorkspace.Initialize").End()
	w.ed = ed
	w.manager.Initialize(host)
	w.manager.NewStage()
	w.pageData.SetupUIData(w.ed.Cache())
	funcs := map[string]func(*document.Element){
		"toggleDimension": w.toggleDimension,
		"inputFilter":     w.contentUI.inputFilter,
		"tagFilter":       w.contentUI.tagFilter,
		"clickFilter":     w.contentUI.clickFilter,
		"dblClickEntry":   w.contentUI.dblClickEntry,
		"hideContent":     w.contentUI.hideContent,
		"showContent":     w.contentUI.showContent,
		"entryDragStart":  w.contentUI.entryDragStart,
		"entryMouseEnter": w.contentUI.entryMouseEnter,
		"entryMouseMove":  w.contentUI.entryMouseMove,
		"entryMouseLeave": w.contentUI.entryMouseLeave,
	}
	funcs = klib.MapJoin(funcs, w.hierarchyUI.setupFuncs())
	funcs = klib.MapJoin(funcs, w.detailsUI.setupFuncs())
	w.CommonWorkspace.InitializeWithUI(host,
		"editor/ui/workspace/stage_workspace.go.html", w.pageData, funcs)
	w.createViewportGrid()
	w.setupCamera()
	w.contentUI.setup(w, w.ed.Events())
	w.hierarchyUI.setup(w)
	w.detailsUI.setup(w)
	w.transformTool.Initialize(host, w, w.ed.History(), &w.ed.Settings().Snapping)
}

func (w *Workspace) Open() {
	defer tracing.NewRegion("StageWorkspace.Open").End()
	w.CommonOpen()
	w.gridShader.Activate()
	w.updateId = w.Host.Updater.AddUpdate(w.update)
	w.contentUI.open()
	w.hierarchyUI.open()
	w.detailsUI.open()
	w.Host.RunOnMainThread(w.Doc.Clean)
}

func (w *Workspace) Close() {
	defer tracing.NewRegion("StageWorkspace.Close").End()
	w.Host.Updater.RemoveUpdate(&w.updateId)
	w.gridShader.Deactivate()
	w.CommonClose()
}

func (w *Workspace) update(deltaTime float64) {
	defer tracing.NewRegion("StageWorkspace.update").End()
	if w.UiMan.IsUpdateDisabled() {
		return
	}
	if !w.contentUI.update(w) {
		return
	}
	if w.IsBlurred || w.UiMan.Group.HasRequests() {
		return
	}
	w.contentUI.processHotkeys(w.Host)
	w.hierarchyUI.processHotkeys(w.Host)
	w.detailsUI.processHotkeys(w.Host)
	if w.camera.Update(w.Host, deltaTime) {
		w.updateGridPosition()
	} else {
		w.processViewportInteractions()
	}
}

func (w *Workspace) updateGridPosition() {
	camPos := w.Host.Camera.Position()
	switch w.camera.Mode() {
	case editor_controls.EditorCameraMode2d:
		w.gridTransform.SetPosition(matrix.NewVec3(
			matrix.Floor(camPos.X()), matrix.Floor(camPos.Y()), 0))
	case editor_controls.EditorCameraMode3d:
		w.gridTransform.SetPosition(matrix.NewVec3(
			matrix.Floor(camPos.X()), 0, matrix.Floor(camPos.Z())))
	}
}

func (w *Workspace) createViewportGrid() {
	defer tracing.NewRegion("StageWorkspace.createViewportGrid").End()
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
	w.gridTransform = matrix.NewTransform(w.Host.WorkGroup())
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	w.gridShader = sd.(*shader_data_registry.ShaderDataGrid)
	w.gridShader.Color = matrix.NewColor(0.5, 0.5, 0.5, 1)
	w.Host.Drawings.AddDrawing(rendering.Drawing{
		Renderer:   w.Host.Window.Renderer,
		Material:   material,
		Mesh:       grid,
		ShaderData: w.gridShader,
		Transform:  &w.gridTransform,
	})
}

func (w *Workspace) setupCamera() {
	defer tracing.NewRegion("StageWorkspace.setupCamera").End()
	w.camera.OnModeChange.Add(func() {
		switch w.camera.Mode() {
		case editor_controls.EditorCameraMode3d:
			// Identity matrix is fine
			w.gridShader.Color.SetA(1)
			w.gridTransform.SetRotation(matrix.Vec3Zero())
		case editor_controls.EditorCameraMode2d:
			w.gridShader.Color.SetA(0)
			w.gridTransform.SetRotation(matrix.NewVec3(90, 0, 0))
		}
		w.updateGridPosition()
	})
	w.camera.SetMode(editor_controls.EditorCameraMode3d, w.Host)
}

func (w *Workspace) toggleDimension(e *document.Element) {
	lbl := e.InnerLabel()
	switch lbl.Text() {
	case "3D":
		lbl.SetText("2D")
		w.camera.SetMode(editor_controls.EditorCameraMode2d, w.Host)
	case "2D":
		lbl.SetText("3D")
		w.camera.SetMode(editor_controls.EditorCameraMode3d, w.Host)
	}
}
