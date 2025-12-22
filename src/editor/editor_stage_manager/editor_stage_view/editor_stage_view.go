/******************************************************************************/
/* editor_stage_view.go                                                       */
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

package editor_stage_view

import (
	"kaiju/editor/editor_controls"
	"kaiju/editor/editor_settings"
	"kaiju/editor/editor_stage_manager"
	"kaiju/editor/editor_stage_manager/data_binding_renderer"
	"kaiju/editor/editor_stage_manager/editor_stage_view/select_tool"
	"kaiju/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaiju/editor/memento"
	"kaiju/editor/project"
	"kaiju/engine"
	"kaiju/engine/assets"
	"kaiju/matrix"
	"kaiju/platform/hid"
	"kaiju/platform/profiler/tracing"
	"kaiju/registry/shader_data_registry"
	"kaiju/rendering"
	"log/slog"
	"weak"
)

type StageView struct {
	host          *engine.Host
	camera        editor_controls.EditorCamera
	gridTransform matrix.Transform
	gridShader    *shader_data_registry.ShaderDataGrid
	manager       editor_stage_manager.StageManager
	transformTool transform_tools.TransformTool
	selectTool    select_tool.SelectTool
}

func (v *StageView) Manager() *editor_stage_manager.StageManager { return &v.manager }

func (v *StageView) Camera() *editor_controls.EditorCamera { return &v.camera }

func (v *StageView) WorkspaceHost() *engine.Host { return v.host }

func (v *StageView) LookAtPoint() matrix.Vec3 { return v.camera.LookAtPoint() }

func (v *StageView) IsView3D() bool { return v.isCamera3D() }

func (v *StageView) Initialize(host *engine.Host, history *memento.History, settings *editor_settings.Settings, editorUI editor_stage_manager.EditorUserInterface) {
	defer tracing.NewRegion("StageView.Initialize").End()
	v.manager.Initialize(host, history, editorUI)
	v.manager.NewStage()
	v.host = host
	v.transformTool.Initialize(host, v, history, &settings.Snapping)
	v.selectTool.Init(host, &v.manager)
	v.createViewportGrid()
	v.setupCamera(&settings.EditorCamera)
	// Data binding visualizers
	weakHost := weak.Make(host)
	v.manager.OnEntitySelected.Add(func(e *editor_stage_manager.StageEntity) {
		data_binding_renderer.Show(weakHost, e)
	})
	v.manager.OnEntityDeselected.Add(func(e *editor_stage_manager.StageEntity) {
		data_binding_renderer.Hide(weakHost, e)
	})
}

func (v *StageView) Open() {
	defer tracing.NewRegion("StageView.Open").End()
	v.gridShader.Activate()
}

func (v *StageView) Close() {
	defer tracing.NewRegion("StageView.Close").End()
	v.gridShader.Deactivate()
}

// Update will update the stage view and return `true` if the view is taking
// control of the keyboard interactions. It'll return false otherwise. If this
// returns true, then the caller shouldn't process any hotkeys or other types
// of keyboard actions.
func (v *StageView) Update(deltaTime float64, proj *project.Project) bool {
	defer tracing.NewRegion("StageView.Update").End()
	v.gridTransform.ResetDirty()
	if v.camera.Update(v.host, deltaTime) {
		v.updateGridPosition()
		v.transformTool.Cancel()
		return true
	} else {
		v.processViewportInteractions()
	}
	kb := &v.host.Window.Keyboard
	if kb.KeyDown(hid.KeyboardKeyDelete) {
		v.manager.DestroySelected()
	} else if kb.HasCtrl() && kb.KeyDown(hid.KeyboardKeyD) {
		v.DuplicateSelected(proj)
		return true
	}
	return false
}

func (v *StageView) SetCameraMode(mode editor_controls.EditorCameraMode) {
	defer tracing.NewRegion("StageView.SetCameraMode").End()
	v.camera.SetMode(mode, v.host)
}

func (v *StageView) updateGridPosition() {
	defer tracing.NewRegion("StageView.updateGridPosition").End()
	camPos := v.host.PrimaryCamera().Position()
	switch v.camera.Mode() {
	case editor_controls.EditorCameraMode2d:
		v.gridTransform.SetPosition(matrix.NewVec3(
			matrix.Floor(camPos.X()), matrix.Floor(camPos.Y()), 0))
	case editor_controls.EditorCameraMode3d:
		v.gridTransform.SetPosition(matrix.NewVec3(
			matrix.Floor(camPos.X()), 0, matrix.Floor(camPos.Z())))
	}
}

func (v *StageView) createViewportGrid() {
	defer tracing.NewRegion("StageView.createViewportGrid").End()
	const gridCount = 100
	const halfGridCount = gridCount / 2
	material, err := v.host.MaterialCache().Material(assets.MaterialDefinitionGrid)
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
	grid := rendering.NewMeshGrid(v.host.MeshCache(), "viewport_grid",
		points, matrix.Color{0.5, 0.5, 0.5, 1})
	v.gridTransform.Initialize(v.host.WorkGroup())
	sd := shader_data_registry.Create(material.Shader.ShaderDataName())
	v.gridShader = sd.(*shader_data_registry.ShaderDataGrid)
	v.gridShader.Color = matrix.NewColor(0.5, 0.5, 0.5, 1)
	v.host.Drawings.AddDrawing(rendering.Drawing{
		Material:   material,
		Mesh:       grid,
		ShaderData: v.gridShader,
		Transform:  &v.gridTransform,
		ViewCuller: &v.host.Cameras.Primary,
	})
	v.gridTransform.ResetDirty()
}

func (v *StageView) setupCamera(settings *editor_settings.EditorCameraSettings) {
	defer tracing.NewRegion("StageView.setupCamera").End()
	v.camera.OnModeChange.Add(func() {
		switch v.camera.Mode() {
		case editor_controls.EditorCameraMode3d:
			// Identity matrix is fine
			v.gridShader.Color.SetA(1)
			v.gridTransform.SetRotation(matrix.Vec3Zero())
		case editor_controls.EditorCameraMode2d:
			v.gridShader.Color.SetA(0)
			v.gridTransform.SetRotation(matrix.NewVec3(90, 0, 0))
		}
		v.updateGridPosition()
	})
	v.camera.SetMode(editor_controls.EditorCameraMode3d, v.host)
	v.camera.Settings = settings
}

func (v *StageView) DuplicateSelected(proj *project.Project) {
	v.manager.DuplicateSelected(proj)
	// The new selection is the duplicated entities
	var callAttachments func(e *editor_stage_manager.StageEntity)
	callAttachments = func(e *editor_stage_manager.StageEntity) {
		for _, de := range e.DataBindings() {
			data_binding_renderer.Attached(de, weak.Make(v.host), &v.manager, e)
		}
		for _, c := range e.Children {
			callAttachments(editor_stage_manager.EntityToStageEntity(c))
		}
	}
	for _, e := range v.manager.HierarchyRespectiveSelection() {
		callAttachments(e)
	}
}
