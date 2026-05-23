/******************************************************************************/
/* editor_stage_view.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package editor_stage_view

import (
	"log/slog"
	"weak"

	"kaijuengine.com/editor/editor_controls"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view/select_tool"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view/transform_tools"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/hid"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/registry/shader_data_registry"
	"kaijuengine.com/rendering"
)

type StageView struct {
	host          *engine.Host
	camera        editor_controls.EditorCamera
	gridTransform matrix.Transform
	gridShader    *shader_data_registry.ShaderDataGrid
	gridVisible   bool
	manager       editor_stage_manager.StageManager
	transformTool transform_tools.TransformTool
	selectTool    select_tool.SelectTool
	transformMan  TransformationManager
	toolOwner     ViewportToolOwner
}

type ViewportToolOwner interface {
	UpdateViewportTool(view *StageView) bool
}

func (v *StageView) Manager() *editor_stage_manager.StageManager { return &v.manager }

func (v *StageView) Camera() *editor_controls.EditorCamera { return &v.camera }

func (v *StageView) WorkspaceHost() *engine.Host { return v.host }

func (v *StageView) LookAtPoint() matrix.Vec3 { return v.camera.LookAtPoint() }

func (v *StageView) IsView3D() bool { return v.isCamera3D() }

func (v *StageView) SetViewportToolOwner(owner ViewportToolOwner) {
	v.toolOwner = owner
}

func (v *StageView) ClearViewportToolOwner(owner ViewportToolOwner) {
	if v.toolOwner == owner {
		v.toolOwner = nil
	}
}

func (v *StageView) Initialize(host *engine.Host, ed EditorStageViewWorkspaceInterface) {
	defer tracing.NewRegion("StageView.Initialize").End()
	v.manager.Initialize(host, ed.History(), ed)
	v.host = host
	v.gridVisible = ed.Settings().ShowGrid
	v.manager.NewStage()
	v.transformTool.Initialize(host, v, ed.History(), &ed.Settings().Snapping)
	v.transformMan.Initialize(v, ed.History(), &ed.Settings().Snapping)
	v.selectTool.Init(host, &v.manager)
	v.createViewportGrid()
	v.applyGridVisibility()
	v.setupCamera(ed)
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
	v.applyGridVisibility()
}

func (v *StageView) Close() {
	defer tracing.NewRegion("StageView.Close").End()
	v.gridShader.Deactivate()
}

// IsGridVisible returns whether the editor viewport grid is currently shown.
func (v *StageView) IsGridVisible() bool { return v.gridVisible }

// SetGridVisible toggles the editor viewport grid. The change is applied
// immediately to the live drawing; persistence is the caller's responsibility.
func (v *StageView) SetGridVisible(visible bool) {
	v.gridVisible = visible
	v.applyGridVisibility()
}

func (v *StageView) applyGridVisibility() {
	if v.gridShader == nil {
		return
	}
	if v.gridVisible {
		v.gridShader.Activate()
	} else {
		v.gridShader.Deactivate()
	}
}

// Update will update the stage view and return `true` if the view is taking
// control of the keyboard interactions. It'll return false otherwise. If this
// returns true, then the caller shouldn't process any hotkeys or other types
// of keyboard actions.
func (v *StageView) Update(deltaTime float64, proj *project.Project) bool {
	defer tracing.NewRegion("StageView.Update").End()
	v.gridTransform.ResetDirty()
	// If we are currently using any of the transformation tools, we shouldn't
	// do any of the other updates like camera
	if v.transformMan.IsBusy() {
		v.transformMan.Update(v.host)
		return true
	}
	if v.camera.Update(v.host, deltaTime) {
		v.updateGridPosition()
		v.transformTool.Cancel()
		v.selectTool.Cancel()
		return true
	} else {
		v.processViewportInteractions()
	}
	kb := &v.host.Window.Keyboard
	if kb.KeyDown(hid.KeyboardKeyDelete) {
		v.manager.DestroySelected()
	} else if kb.HasCtrlOrMeta() && kb.KeyDown(hid.KeyboardKeyD) {
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
	cam := v.host.PrimaryCamera()
	camPos := cam.Position()
	switch v.camera.Mode() {
	case editor_controls.EditorCameraMode2d:
		v.gridTransform.SetPosition(matrix.NewVec3(
			matrix.Floor(camPos.X()), matrix.Floor(camPos.Y()), -cam.FarPlane()*0.45))
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

func (v *StageView) setupCamera(ed EditorStageViewWorkspaceInterface) {
	defer tracing.NewRegion("StageView.setupCamera").End()
	pjs := &ed.Project().Settings
	v.camera.OnModeChange.Add(func(mode editor_controls.EditorCameraMode) {
		switch mode {
		case editor_controls.EditorCameraMode3d:
			// Identity matrix is fine
			v.gridShader.Color.SetA(1)
			v.gridTransform.SetRotation(matrix.Vec3Zero())
			pjs.EditorSettings.CameraMode = editor_controls.EditorCameraMode3d
		case editor_controls.EditorCameraMode2d:
			v.gridShader.Color.SetA(0)
			v.gridTransform.SetRotation(matrix.NewVec3(90, 0, 0))
			pjs.EditorSettings.CameraMode = editor_controls.EditorCameraMode2d
		}
		v.updateGridPosition()
		if err := pjs.Save(ed.ProjectFileSystem()); err != nil {
			slog.Error("there was an error saving the project settings during setupCamera", "error", err)
		}
	})
	v.camera.SetMode(pjs.EditorSettings.CameraMode, v.host)
	v.camera.Settings = &ed.Settings().EditorCamera
}

func (v *StageView) DuplicateSelected(proj *project.Project) {
	if !v.manager.HasSelection() {
		return
	}
	v.manager.DuplicateSelected(proj)
	// The new selection is the duplicated entities
	weakHost := weak.Make(v.host)
	var callAttachments func(e *editor_stage_manager.StageEntity)
	callAttachments = func(e *editor_stage_manager.StageEntity) {
		for _, de := range e.DataBindings() {
			data_binding_renderer.Attached(de, weakHost, &v.manager, e)
			if v.manager.IsSelected(e) {
				data_binding_renderer.ShowSpecific(de, weakHost, e)
			}
		}
		for _, c := range e.Children {
			callAttachments(editor_stage_manager.EntityToStageEntity(c))
		}
	}
	for _, e := range v.manager.HierarchyRespectiveSelection() {
		callAttachments(e)
	}
	v.transformMan.EnableTranslationTool()
}
